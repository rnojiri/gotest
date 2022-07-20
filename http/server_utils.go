package http

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"
)

/**
* Common helper functions.
* @author rnojiri
**/

// DoRequest - does a request
func (hs *Server) DoRequest(request *Request) *http.Response {

	transportCore := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: transportCore,
		Timeout:   10 * time.Second,
	}

	req, err := http.NewRequest(
		request.Method,
		fmt.Sprintf("http://%s:%d/%s", hs.configuration.Host, hs.configuration.Port, request.URI),
		bytes.NewBuffer(request.Body),
	)
	if err != nil {
		hs.configuration.T.Fatalf("error creating a new request: %v", err)
	}

	req.Header = request.Headers

	res, err := client.Do(req)
	if err != nil {
		hs.configuration.T.Fatalf("error executing request: %v", err)
	}

	return res
}

// WaitForServerRequest - wait until timeout or for the server sets the request in the channel
func WaitForServerRequest(server *Server, waitFor, maxRequestTimeout time.Duration) *Request {

	var request *Request
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

// AddHeaders - copy all the headers
func AddHeaders(dest http.Header, source http.Header) {

	for header, valueList := range source {

		for _, v := range valueList {
			dest.Add(header, v)
		}
	}
}

// CleanURI - cleans and validates the URI
func CleanURI(name string) string {

	if !strings.HasPrefix(name, "/") {
		name += "/"
	}

	return multipleBarRegexp.ReplaceAllString(name, "/")
}
