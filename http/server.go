package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/jinzhu/copier"
)

/**
* Mocks a http server and offers a way to validate the sent content.
* @author rnojiri
**/

// Request - the request data made to the server
type Request struct {
	URI     string
	Body    []byte
	Method  string
	Headers http.Header
}

// Response - the endpoint response data
type Response struct {
	// Body - the body to be put in the response
	Body interface{}
	// Headers - the headers to be included in the response
	Headers http.Header
	// Status - the code status to be returned
	Status int
	// Wait - a time to wait until responds
	Wait time.Duration
}

// Endpoint - an endpoint to be listened
type Endpoint struct {
	// URI - the endpoint's uri
	URI string
	// QueryString - the query string format
	QueryString []string
	// Methods - the list of http methods (GET, POST, ...) containing the respective response
	Methods map[string]Response
	// Regexp - activates regular expression for uris
	Regexp bool
}

// Server - the server listening for HTTP requests
type Server struct {
	server        *httptest.Server
	requests      []Request
	responseMap   map[string]map[string]Endpoint
	errors        []error
	configuration *Configuration
	mode          string
	mutex         sync.Mutex
}

// Configuration - configuration
type Configuration struct {
	Host      string
	Port      int
	Responses map[string][]Endpoint
	T         *testing.T
}

var multipleBarRegexp = regexp.MustCompile("[/]+")

// NewServer - creates a new HTTP listener server
func NewServer(configuration *Configuration) *Server {

	if configuration == nil {
		panic(fmt.Errorf("null configuration"))
	}

	if len(configuration.Responses) == 0 {
		panic(fmt.Errorf("expected at least one response"))
	}

	hs := &Server{
		requests: []Request{},
	}

	hs.responseMap = map[string]map[string]Endpoint{}
	for mode, responses := range configuration.Responses {

		hs.responseMap[mode] = map[string]Endpoint{}
		for _, response := range responses {
			response.URI = CleanURI(response.URI)
			hs.responseMap[mode][response.URI] = response
		}

		hs.mode = mode
	}

	hs.server = httptest.NewUnstartedServer(http.HandlerFunc(hs.handler))

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", configuration.Host, configuration.Port))
	if err != nil {
		panic(err)
	}

	confCopy := Configuration{}
	copier.Copy(&confCopy, configuration)

	hs.server.Listener = listener
	hs.mutex = sync.Mutex{}
	hs.server.Start()
	hs.configuration = &confCopy

	return hs
}

// handler - handles all requests
func (hs *Server) handler(res http.ResponseWriter, req *http.Request) {

	cleanURI := CleanURI(req.RequestURI)

	modeMaps, ok := hs.responseMap[hs.mode]
	if !ok {
		hs.configuration.T.Fatalf("no configuration set with name: %s", hs.mode)
	}

	var endpoint Endpoint
	var err error
	found := false

	for uri, item := range modeMaps {

		match := false

		if item.Regexp {

			match, err = regexp.MatchString(uri, cleanURI)
			if err != nil {
				hs.configuration.T.Fatalf("failed to run regexp: %s", cleanURI)
			}

			if match {
				endpoint = item
				found = true
				break
			}
		}

		if !match && uri == cleanURI {
			endpoint = item
			found = true
			break
		}
	}

	if !found {
		hs.configuration.T.Fatalf("no enpoint configured with uri: %s", cleanURI)
	}

	response, ok := endpoint.Methods[req.Method]
	if !ok {
		hs.configuration.T.Fatalf("no method configured under uri: %s", cleanURI)
		return
	}

	if response.Wait != 0 {
		time.Sleep(response.Wait)
	}

	AddHeaders(res.Header(), response.Headers)

	res.WriteHeader(response.Status)

	if response.Body != nil {

		var inBytes []byte
		var err error
		switch response.Body.(type) {
		case int:
			inBytes = []byte(strconv.FormatInt(int64(response.Body.(int)), 10))
		case string:
			inBytes = []byte(response.Body.(string))
		case bool:
			inBytes = []byte(strconv.FormatBool(response.Body.(bool)))
		default:
			inBytes, err = json.Marshal(response.Body)
			if err != nil {
				hs.configuration.T.Fatalf("error marshaling json: %v", err)
			}
		}

		_, err = res.Write(inBytes)
		if err != nil {
			hs.errors = append(hs.errors, err)
			return
		}
	}

	bufferReqBody := new(bytes.Buffer)
	_, err = bufferReqBody.ReadFrom(req.Body)
	if err != nil {
		hs.configuration.T.Fatalf("error reading request body: %v", err)
	}

	hs.mutex.Lock()
	hs.mutex.Unlock()

	hs.requests = append(
		hs.requests,
		Request{
			URI:     cleanURI,
			Body:    bufferReqBody.Bytes(),
			Headers: req.Header.Clone(),
			Method:  req.Method,
		},
	)
}

// Close - closes this server
func (hs *Server) Close() {

	if hs.server != nil {
		hs.server.Close()
	}
}

// RequestChannel - reads from the request channel
func (hs *Server) RequestChannel() []Request {

	return hs.requests
}

func (hs *Server) FirstRequest() *Request {

	hs.mutex.Lock()
	hs.mutex.Unlock()

	if len(hs.requests) == 0 {
		return nil
	}

	req := hs.requests[0]

	if len(hs.requests) == 0 {
		hs.requests = []Request{}
	} else {
		hs.requests = hs.requests[1:]
	}

	return &req
}

// SetMode - sets the server mode
func (hs *Server) SetMode(mode string) {

	hs.mode = mode
}
