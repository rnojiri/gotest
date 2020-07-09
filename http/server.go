package http

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"regexp"
	"time"

	"github.com/jinzhu/copier"
)

/**
* Mocks a http server and offers a way to validate the sent content.
* @author rnojiri
**/

// RequestData - the request data sent to the server
type RequestData struct {
	URI     string
	Body    string
	Method  string
	Headers http.Header
	Date    time.Time
	Host    string
	Port    int
}

// ResponseData - the expected response data for each configured URI
type ResponseData struct {
	RequestData
	Status int
}

// Server - the server listening for HTTP requests
type Server struct {
	server         *httptest.Server
	requestChannel chan *RequestData
	responseMap    map[string]map[string]ResponseData
	errors         []error
	configuration  *Configuration
	mode           string
}

// Configuration - configuration
type Configuration struct {
	Host        string
	Port        int
	ChannelSize int
	Responses   map[string][]ResponseData
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
		requestChannel: make(chan *RequestData, configuration.ChannelSize),
	}

	hs.responseMap = map[string]map[string]ResponseData{}
	for mode, responses := range configuration.Responses {

		hs.responseMap[mode] = map[string]ResponseData{}
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
	hs.server.Start()
	hs.configuration = &confCopy

	return hs
}

// handler - handles all requests
func (hs *Server) handler(res http.ResponseWriter, req *http.Request) {

	cleanURI := CleanURI(req.RequestURI)

	modeMaps, ok := hs.responseMap[hs.mode]
	if !ok {
		hs.errors = append(hs.errors, fmt.Errorf("no mode named: %s", hs.mode))
		res.WriteHeader(http.StatusFailedDependency)
		return
	}

	responseData, ok := modeMaps[cleanURI]
	if !ok || responseData.Method != req.Method {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	combinedHeaders := http.Header{}
	CopyHeaders(responseData.Headers, &combinedHeaders)

	if responseData.Status != http.StatusOK {
		res.WriteHeader(responseData.Status)
	}

	if len(responseData.Body) > 0 {
		_, err := res.Write([]byte(responseData.Body))
		if err != nil {
			hs.errors = append(hs.errors, err)
			return
		}
	}

	bufferReqBody := new(bytes.Buffer)
	bufferReqBody.ReadFrom(req.Body)

	hs.requestChannel <- &RequestData{
		URI:     cleanURI,
		Body:    bufferReqBody.String(),
		Headers: req.Header,
		Method:  req.Method,
		Date:    time.Now(),
		Host:    hs.configuration.Host,
		Port:    hs.configuration.Port,
	}
}

// Close - closes this server
func (hs *Server) Close() {

	if hs.server != nil {
		hs.server.Close()
	}
}

// RequestChannel - reads from the request channel
func (hs *Server) RequestChannel() <-chan *RequestData {

	return hs.requestChannel
}

// SetMode - sets the server mode
func (hs *Server) SetMode(mode string) {

	hs.mode = mode
}
