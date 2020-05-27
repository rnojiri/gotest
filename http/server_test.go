package http_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	gotesthttp "github.com/uol/gotest/http"
)

/**
* The tests for the http server used by tests.
* @author rnojiri
**/

const (
	testHost    string = "localhost"
	testPort    int    = 18080
	channelSize int    = 5
)

// createDummyResponse - creates a dummy response data
func createDummyResponse() gotesthttp.ResponseData {

	headers := http.Header{}
	headers.Add("Content-type", "text/plain; charset=utf-8")

	return gotesthttp.ResponseData{
		RequestData: gotesthttp.RequestData{
			URI:     "/test",
			Body:    "test body",
			Method:  "GET",
			Headers: headers,
		},
		Status: http.StatusOK,
	}
}

// Test404 - tests when a non mapped response is called
func Test404(t *testing.T) {

	server := gotesthttp.NewServer(testHost, testPort, channelSize, []gotesthttp.ResponseData{createDummyResponse()})
	defer server.Close()

	response := gotesthttp.DoRequest(testHost, testPort, &gotesthttp.RequestData{
		URI:    "/not",
		Method: "GET",
	})

	assert.Equal(t, http.StatusNotFound, response.Status, "expected 404 status")

	response = gotesthttp.DoRequest(testHost, testPort, &gotesthttp.RequestData{
		URI:    "/test",
		Method: "POST",
	})

	assert.Equal(t, http.StatusNotFound, response.Status, "expected 404 status")

	response = gotesthttp.DoRequest(testHost, testPort, &gotesthttp.RequestData{
		URI:    "/test",
		Method: "GET",
	})

	assert.Equal(t, http.StatusOK, response.Status, "expected 200 status")
}

// TestSuccess - tests when everything goes right
func TestSuccess(t *testing.T) {

	configuredResponse := createDummyResponse()

	server := gotesthttp.NewServer(testHost, testPort, channelSize, []gotesthttp.ResponseData{configuredResponse})
	defer server.Close()

	reqHeader := http.Header{}
	reqHeader.Add("Content-type", "text/plain; charset=utf-8")

	clientRequest := &gotesthttp.RequestData{
		URI:     "/test",
		Body:    "test body",
		Method:  "GET",
		Headers: reqHeader,
	}

	serverResponse := gotesthttp.DoRequest(testHost, testPort, clientRequest)
	if !compareResponses(t, &configuredResponse, serverResponse) {
		return
	}

	serverRequest := gotesthttp.WaitForServerRequest(server, time.Second, 10*time.Second)
	compareRequests(t, clientRequest, serverRequest)
}

// TestMultipleResponses - tests when everything goes right with multiple responses
func TestMultipleResponses(t *testing.T) {

	configuredResponse1 := createDummyResponse()
	configuredResponse1.URI = "/text"
	configuredResponse1.Method = "POST"

	configuredResponse2 := createDummyResponse()
	configuredResponse2.URI = "/json"
	configuredResponse2.Method = "PUT"
	configuredResponse2.Status = http.StatusCreated
	configuredResponse2.Body = `{"metric": "test-metric", "value": 1.0}`
	configuredResponse2.Headers.Del("Content-type")
	configuredResponse2.Headers.Set("Content-type", "application/json")

	server := gotesthttp.NewServer(testHost, testPort, channelSize, []gotesthttp.ResponseData{configuredResponse1, configuredResponse2})
	defer server.Close()

	reqHeader1 := http.Header{}
	reqHeader1.Set("Content-type", "text/plain; charset=utf-8")

	clientRequest1 := &gotesthttp.RequestData{
		URI:     "/text",
		Body:    "some text",
		Method:  "POST",
		Headers: reqHeader1,
	}

	serverResponse := gotesthttp.DoRequest(testHost, testPort, clientRequest1)
	if !compareResponses(t, &configuredResponse1, serverResponse) {
		return
	}

	serverRequest := gotesthttp.WaitForServerRequest(server, time.Second, 10*time.Second)
	compareRequests(t, clientRequest1, serverRequest)

	reqHeader2 := http.Header{}
	reqHeader2.Set("Content-type", "application/json")

	clientRequest2 := &gotesthttp.RequestData{
		URI:     "/json",
		Body:    `{"metric": "test-metric", "value": 1.0}`,
		Method:  "PUT",
		Headers: reqHeader2,
	}

	serverResponse = gotesthttp.DoRequest(testHost, testPort, clientRequest2)
	if !compareResponses(t, &configuredResponse2, serverResponse) {
		return
	}

	serverRequest = gotesthttp.WaitForServerRequest(server, time.Second, 10*time.Second)
	compareRequests(t, clientRequest2, serverRequest)
}

// compareResponses - compares two responses
func compareResponses(t *testing.T, r1 *gotesthttp.ResponseData, r2 *gotesthttp.ResponseData) bool {

	result := true

	result = result && assert.Equal(t, r1.Body, r2.Body, "same body expected")
	result = result && containsHeaders(t, r1.Headers, r2.Headers)
	result = result && assert.Equal(t, r1.Method, r2.Method, "same method expected")
	result = result && assert.Equal(t, r1.Status, r2.Status, "same status expected")
	result = result && assert.Equal(t, r1.URI, r2.URI, "same URI expected")

	return result
}

// compareRequests - compares two requests
func compareRequests(t *testing.T, r1 *gotesthttp.RequestData, r2 *gotesthttp.RequestData) bool {

	result := true

	result = result && assert.Equal(t, r1.Body, r2.Body, "same body expected")
	result = result && containsHeaders(t, r1.Headers, r2.Headers)
	result = result && assert.Equal(t, r1.Method, r2.Method, "same method expected")
	result = result && assert.Equal(t, r1.URI, r2.URI, "same URI expected")

	return result
}

// containsHeaders - checks for the headers
func containsHeaders(t *testing.T, mustExist, fullSet http.Header) bool {

	if mustExist == nil {
		return true
	}

	assert.NotNil(t, fullSet, "the full set of headers must not be null")

	for mustExistHeader, mustExistValues := range mustExist {

		if !assert.Truef(t, len(fullSet[mustExistHeader]) > 0, "expected a list of values for the header: %s", mustExistHeader) {
			return false
		}

		if !assert.Equal(t, fullSet[mustExistHeader], mustExistValues, "expected some headers") {
			return false
		}
	}

	return true
}