package http_test

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	randomdata "github.com/Pallinder/go-randomdata"
	"github.com/stretchr/testify/assert"
	gotesthttp "github.com/uol/gotest/http"
)

/**
* The tests for the http server used by tests.
* @author rnojiri
**/

var defaultConf gotesthttp.Configuration = gotesthttp.Configuration{
	Host:        "localhost",
	Port:        18080,
	ChannelSize: 5,
}

// createDummyEndpoint - creates a dummy response data
func createDummyEndpoint(method string) gotesthttp.Endpoint {

	headers := http.Header{}
	headers.Add("Content-type", "text/plain; charset=utf-8")
	headers.Add("X-custom", randomdata.Adjective())

	return gotesthttp.Endpoint{
		URI: "/" + strings.ToLower(randomdata.SillyName()),
		Methods: map[string]gotesthttp.Response{
			method: {
				Body:    randomdata.City(),
				Headers: headers,
				Status:  http.StatusOK,
			},
		},
	}
}

func randomMethod() string {
	return randomdata.StringSample("GET", "POST", "PUT")
}

func createRequestFromEndpoint(method string, endpoint *gotesthttp.Endpoint) *gotesthttp.Request {

	return &gotesthttp.Request{
		URI:     endpoint.URI,
		Body:    []byte(endpoint.Methods[method].Body.(string)),
		Method:  method,
		Headers: endpoint.Methods[method].Headers,
	}
}

// TestSuccess - tests when everything goes right
func TestSuccess(t *testing.T) {

	method := randomMethod()
	endpoint := createDummyEndpoint(method)

	defaultConf.T = t
	defaultConf.Responses = map[string][]gotesthttp.Endpoint{
		"default": {endpoint},
	}

	server := gotesthttp.NewServer(&defaultConf)
	defer server.Close()

	clientRequest := createRequestFromEndpoint(method, &endpoint)

	serverResponse := server.DoRequest(clientRequest)
	if !compareResponses(t, defaultConf.Responses["default"][0].Methods[method], serverResponse) {
		return
	}

	serverRequest := gotesthttp.WaitForServerRequest(server, time.Second, 10*time.Second)
	compareRequests(t, clientRequest, serverRequest)
}

// TestMultipleResponses - tests when everything goes right with multiple responses
func TestMultipleResponses(t *testing.T) {

	r1Method := randomMethod()
	r2Method := randomMethod()

	endpoint1 := createDummyEndpoint(r1Method)
	endpoint1.URI = "/text"

	endpoint2 := gotesthttp.Endpoint{
		URI: "/json",
		Methods: map[string]gotesthttp.Response{
			r2Method: {
				Status: http.StatusCreated,
				Body:   `{"metric": "test-metric", "value": 1.0}`,
				Headers: http.Header{
					"X-Test": []string{"some"},
				},
			},
		},
	}

	defaultConf.T = t
	defaultConf.Responses = map[string][]gotesthttp.Endpoint{
		"default": {endpoint1, endpoint2},
	}

	server := gotesthttp.NewServer(&defaultConf)
	defer server.Close()

	clientRequest1 := createRequestFromEndpoint(r1Method, &endpoint1)

	serverResponse := server.DoRequest(clientRequest1)
	if !compareResponses(t, endpoint1.Methods[r1Method], serverResponse) {
		return
	}

	serverRequest := gotesthttp.WaitForServerRequest(server, time.Second, 10*time.Second)
	compareRequests(t, clientRequest1, serverRequest)

	clientRequest2 := createRequestFromEndpoint(r2Method, &endpoint2)

	serverResponse = server.DoRequest(clientRequest2)
	if !compareResponses(t, endpoint2.Methods[r2Method], serverResponse) {
		return
	}

	serverRequest = gotesthttp.WaitForServerRequest(server, time.Second, 10*time.Second)
	compareRequests(t, clientRequest2, serverRequest)
}

// compareResponses - compares two responses
func compareResponses(t *testing.T, r1 gotesthttp.Response, r2 *http.Response) bool {

	result := true

	receivedBody, err := ioutil.ReadAll(r2.Body)
	assert.NoError(t, err, "expects no error reading the response body")

	result = result && assert.Equal(t, r1.Body.(string), string(receivedBody), "same body expected")
	result = result && containsHeaders(t, r1.Headers, r2.Header)
	result = result && assert.Equal(t, r1.Status, r2.StatusCode, "same status expected")

	return result
}

// compareRequests - compares two requests
func compareRequests(t *testing.T, r1 *gotesthttp.Request, r2 *gotesthttp.Request) bool {

	result := true

	result = result && assert.Equal(t, r1.Body, r2.Body, "same body expected")
	result = result && containsHeaders(t, r1.Headers, r2.Headers)
	result = result && assert.Equal(t, r1.Method, r2.Method, "same method expected")
	result = result && assert.Equal(t, r1.URI, r2.URI, "same URI expected")

	return result
}

// containsHeaders - checks for the headers
func containsHeaders(t *testing.T, mustExist, responseheaders http.Header) bool {

	if mustExist == nil {
		return true
	}

	assert.NotNil(t, responseheaders, "the response set of headers must not be null")

	for mustExistHeader, mustExistValues := range mustExist {

		list, ok := responseheaders[mustExistHeader]

		if !assert.Truef(t, ok, "expects the header to exist: %s", mustExistHeader) {
			return false
		}

		if !assert.ElementsMatch(t, mustExistValues, list, "expected same headers") {
			return false
		}
	}

	return true
}

// TestSuccessMultiModes - tests when everything goes right
func TestSuccessMultiModes(t *testing.T) {

	r1Method := randomMethod()
	r2Method := randomMethod()

	endpoint1 := createDummyEndpoint(r1Method)
	endpoint2 := createDummyEndpoint(r2Method)

	defaultConf.T = t
	defaultConf.Responses = map[string][]gotesthttp.Endpoint{
		"mode1": {endpoint1},
		"mode2": {endpoint2},
	}

	server := gotesthttp.NewServer(&defaultConf)
	defer server.Close()

	reqHeader := http.Header{}
	reqHeader.Add("Content-type", "text/plain; charset=utf-8")

	clientRequest1 := createRequestFromEndpoint(r1Method, &endpoint1)
	clientRequest2 := createRequestFromEndpoint(r2Method, &endpoint2)

	server.SetMode("mode2")

	serverResponse := server.DoRequest(clientRequest2)
	if !compareResponses(t, endpoint2.Methods[r2Method], serverResponse) {
		return
	}

	serverRequest := gotesthttp.WaitForServerRequest(server, time.Second, 10*time.Second)
	compareRequests(t, clientRequest2, serverRequest)

	server.SetMode("mode1")

	serverResponse = server.DoRequest(clientRequest1)
	if !compareResponses(t, endpoint1.Methods[r1Method], serverResponse) {
		return
	}

	serverRequest = gotesthttp.WaitForServerRequest(server, time.Second, 10*time.Second)
	compareRequests(t, clientRequest1, serverRequest)
}

// TestWaitResponse - tests the wait parameter
func TestWaitResponse(t *testing.T) {

	randomMillis := randomdata.Number(1, 10)

	method := randomMethod()
	endpoint := createDummyEndpoint(method)
	m := endpoint.Methods[method]
	m.Wait = time.Duration(randomMillis*100) * time.Millisecond
	endpoint.Methods[method] = m

	defaultConf.T = t
	defaultConf.Responses = map[string][]gotesthttp.Endpoint{
		"default": {endpoint},
	}

	server := gotesthttp.NewServer(&defaultConf)
	defer server.Close()

	clientRequest := createRequestFromEndpoint(method, &endpoint)

	start := time.Now()

	serverResponse := server.DoRequest(clientRequest)
	if !compareResponses(t, endpoint.Methods[method], serverResponse) {
		return
	}

	requestTime := time.Since(start)

	if !assert.InDelta(t, int64(randomMillis*100), requestTime.Milliseconds(), 5, "expected same amount of time") {
		return
	}

	serverRequest := gotesthttp.WaitForServerRequest(server, time.Duration(randomMillis+1)*time.Second, 10*time.Second)
	compareRequests(t, clientRequest, serverRequest)
}

// TestURIRegexp - tests the uri regexp
func TestURIRegexp(t *testing.T) {

	r1Method := randomMethod()
	r2Method := randomMethod()

	endpoint1 := createDummyEndpoint(r1Method)
	endpoint1.URI = "/normal"

	endpoint2 := gotesthttp.Endpoint{
		URI: "/regexp[0-9]+",
		Methods: map[string]gotesthttp.Response{
			r2Method: {
				Status: http.StatusOK,
				Body:   "ok",
			},
		},
		Regexp: true,
	}

	defaultConf.T = t
	defaultConf.Responses = map[string][]gotesthttp.Endpoint{
		"default": {endpoint1, endpoint2},
	}

	server := gotesthttp.NewServer(&defaultConf)
	defer server.Close()

	clientRequest1 := createRequestFromEndpoint(r1Method, &endpoint1)

	serverResponse := server.DoRequest(clientRequest1)
	if !compareResponses(t, endpoint1.Methods[r1Method], serverResponse) {
		return
	}

	serverRequest := gotesthttp.WaitForServerRequest(server, time.Second, 10*time.Second)
	compareRequests(t, clientRequest1, serverRequest)

	clientRequest2 := createRequestFromEndpoint(r2Method, &endpoint2)
	clientRequest2.URI = "/regexp5"

	serverResponse = server.DoRequest(clientRequest2)
	if !compareResponses(t, endpoint2.Methods[r2Method], serverResponse) {
		return
	}

	serverRequest = gotesthttp.WaitForServerRequest(server, time.Second, 10*time.Second)
	compareRequests(t, clientRequest2, serverRequest)
}
