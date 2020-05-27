package http

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"regexp"
	"time"
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
	responseMap    map[string]ResponseData
	errors         []error
}

var multipleBarRegexp = regexp.MustCompile("[/]+")

// NewServer - creates a new HTTP listener server
func NewServer(host string, port, channelSize int, responses []ResponseData) *Server {

	if len(responses) == 0 {
		panic(fmt.Errorf("expected at least one response"))
	}

	hs := &Server{
		requestChannel: make(chan *RequestData, channelSize),
	}

	hs.responseMap = map[string]ResponseData{}
	for _, response := range responses {
		response.URI = CleanURI(response.URI)
		hs.responseMap[response.URI] = response
	}

	hs.server = httptest.NewUnstartedServer(http.HandlerFunc(hs.handler))

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		panic(err)
	}

	hs.server.Listener = listener
	hs.server.Start()

	return hs
}

// handler - handles all requests
func (hs *Server) handler(res http.ResponseWriter, req *http.Request) {

	cleanURI := CleanURI(req.RequestURI)

	responseData, ok := hs.responseMap[cleanURI]
	if !ok || responseData.Method != req.Method {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	combinedHeaders := res.Header()

	CopyHeaders(responseData.Headers, combinedHeaders)
	CopyHeaders(req.Header, combinedHeaders)

	res.WriteHeader(responseData.Status)

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
