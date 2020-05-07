package gotest

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/uol/funks"
)

/**
* Common helper functions.
* @author rnojiri
**/

// CreateNewTestHTTPServer - creates a new server
func CreateNewTestHTTPServer(testServerHost string, testServerPort int, responses []ResponseData) *HTTPServer {

	s, err := NewHTTPServer(testServerHost, testServerPort, 5, responses)
	if err != nil {
		panic(err)
	}

	return s
}

// DoRequest - does a request
func DoRequest(testServerHost string, testServerPort int, request *RequestData) *ResponseData {

	client := funks.CreateHTTPClient(time.Second, true)

	req, err := http.NewRequest(request.Method, fmt.Sprintf("http://%s:%d/%s", testServerHost, testServerPort, request.URI), bytes.NewBuffer([]byte(request.Body)))
	if err != nil {
		panic(err)
	}

	if len(request.Headers) > 0 {
		CopyHeaders(request.Headers, req.Header)
	}

	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	result, err := ParseResponse(res, request.Date)
	if err != nil {
		panic(err)
	}

	result.URI = request.URI

	return result
}

// WaitForHTTPServerRequest - wait until timeout or for the server sets the request in the channel
func WaitForHTTPServerRequest(server *HTTPServer, waitFor, maxRequestTimeout time.Duration) *RequestData {

	var request *RequestData
	start := time.Now()

	for {
		select {
		case request = <-server.RequestChannel():
		default:
		}

		if request != nil {
			break
		}

		<-time.After(waitFor)

		if time.Since(start) > maxRequestTimeout {
			break
		}
	}

	return request
}
