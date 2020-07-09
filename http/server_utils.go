package http

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/uol/funks"
)

/**
* Common helper functions.
* @author rnojiri
**/

// DoRequest - does a request
func DoRequest(testServerHost string, testServerPort int, request *RequestData) *ResponseData {

	client := funks.CreateHTTPClient(time.Second, true)

	req, err := http.NewRequest(request.Method, fmt.Sprintf("http://%s:%d/%s", testServerHost, testServerPort, request.URI), bytes.NewBuffer([]byte(request.Body)))
	if err != nil {
		panic(err)
	}

	if len(request.Headers) > 0 {
		CopyHeaders(request.Headers, &req.Header)
	}

	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	result, err := parseResponse(res, request.Date)
	if err != nil {
		panic(err)
	}

	result.URI = request.URI

	return result
}

// parseResponse - parses the response using the local struct as result
func parseResponse(res *http.Response, reqDate time.Time) (*ResponseData, error) {

	bufferReqBody := new(bytes.Buffer)
	_, err := bufferReqBody.ReadFrom(res.Body)
	if err != nil {
		return nil, err
	}

	host, portStr, err := net.SplitHostPort(res.Request.Host)
	if err != nil {
		return nil, err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}

	return &ResponseData{
		RequestData: RequestData{
			URI:     res.Request.RequestURI,
			Body:    bufferReqBody.String(),
			Headers: res.Header,
			Method:  res.Request.Method,
			Date:    reqDate,
			Host:    host,
			Port:    port,
		},
		Status: res.StatusCode,
	}, nil
}

// WaitForServerRequest - wait until timeout or for the server sets the request in the channel
func WaitForServerRequest(server *Server, waitFor, maxRequestTimeout time.Duration) *RequestData {

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

// CopyHeaders - copy all the headers
func CopyHeaders(source http.Header, dest *http.Header) {

	if len(source) > 0 {
		for header, valueList := range source {
			for _, v := range valueList {
				dest.Set(header, v)
			}
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
