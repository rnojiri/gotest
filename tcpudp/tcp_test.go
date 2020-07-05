package tcpudp_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	tcpudp "github.com/uol/gotest/tcpudp"
	gotest "github.com/uol/gotest/utils"
)

//
// Tests for tcp server.
// author: rnojiri
//

const (
	testHost   string = "localhost"
	numMsgChan int    = 100
	bufferSize int    = 256
)

var (
	defaultTCPConf tcpudp.TCPConfiguration = tcpudp.TCPConfiguration{
		ServerConfiguration: tcpudp.ServerConfiguration{
			Host:               testHost,
			MessageChannelSize: numMsgChan,
			ReadBufferSize:     bufferSize,
		},
		ReadTimeout: time.Second,
	}
)

// TestTCPCreateServer - tests creating the server only (not accepting connections)
func TestTCPCreateServer(t *testing.T) {

	s, p := tcpudp.NewTCPServer(&defaultTCPConf, false)

	if !assert.NotNil(t, s, "expected a valid instance") {
		return
	}

	assert.GreaterOrEqual(t, p, 10000, "expected port greater than 10000")
}

func mustCreateTCPServer(autoStart bool, readTimeout time.Duration) (*tcpudp.TCPServer, int) {

	s, port := tcpudp.NewTCPServer(&defaultTCPConf, autoStart)

	return s, port
}

// TestTCPNotStartedServerStop - tests the stop server function
func TestTCPNotStartedServerStop(t *testing.T) {

	s, _ := mustCreateTCPServer(false, time.Second)
	if s == nil {
		return
	}

	err := s.Stop()
	assert.NoError(t, err, "error not expected")
}

// TestTCPStartedServerStop - tests the stop server function
func TestTCPStartedServerStop(t *testing.T) {

	s, _ := mustCreateTCPServer(true, time.Second)
	if s == nil {
		return
	}

	err := s.Stop()
	assert.NoError(t, err, "error not expected")
}

// TestTCPOneMessage - tests the server with only one message
func TestTCPOneMessage(t *testing.T) {

	s, port := mustCreateTCPServer(true, time.Second)
	defer s.Stop()
	if s == nil {
		return
	}

	conn, err := tcpudp.ConnectTCP(testHost, port, time.Second)
	if !assert.NoError(t, err, "expected no error connecting") {
		return
	}

	payload := "test"

	err = tcpudp.WriteTCP(conn, payload)
	if !assert.NoError(t, err, "expected no error writing") {
		return
	}

	message := <-s.MessageChannel()
	if !assert.Equal(t, payload, message.Message, "expected same value") {
		return
	}

	assert.Len(t, s.GetErrors(), 0, "expected no errors")
}

// TestTCPMultipleMessages - tests the server with multiple messages
func TestTCPMultipleMessages(t *testing.T) {

	s, port := mustCreateTCPServer(true, time.Second)
	defer s.Stop()
	if s == nil {
		return
	}

	conn, err := tcpudp.ConnectTCP(testHost, port, time.Second)
	if !assert.NoError(t, err, "expected no error connecting") {
		return
	}

	messageFormat := "test%d\n"
	numMessages := gotest.RandomInt(2, 10)
	for i := 0; i < numMessages; i++ {

		payload := fmt.Sprintf(messageFormat, i)

		err = tcpudp.WriteTCP(conn, payload)
		if !assert.NoError(t, err, "expected no error writing") {
			return
		}
	}

	message := <-s.MessageChannel()

	messages := strings.Split(message.Message, "\n")
	messages = messages[:len(messages)-1] //removes the last blank item
	if !assert.Len(t, messages, numMessages, "expected same number of messages") {
		return
	}

	for i := 0; i < numMessages; i++ {
		expected := fmt.Sprintf(messageFormat, i)
		if !assert.Equal(t, expected, messages[i]+"\n", "expected same message") {
			return
		}
	}

	assert.Len(t, s.GetErrors(), 0, "expected no errors")
}

// TestTCPServerResponse - tests the server mocked response
func TestTCPServerResponse(t *testing.T) {

	responseConf := defaultTCPConf
	responseConf.ResponseString = "response"
	responseConf.WriteTimeout = time.Second

	s, port := tcpudp.NewTCPServer(&responseConf, true)
	defer s.Stop()
	if s == nil {
		return
	}

	conn, err := tcpudp.ConnectTCP(testHost, port, 3*time.Second)
	if !assert.NoError(t, err, "expected no error connecting") {
		return
	}

	payload := "request"

	err = tcpudp.WriteTCP(conn, payload)
	if !assert.NoError(t, err, "expected no error writing") {
		return
	}

	response, err := tcpudp.ReadTCP(conn, bufferSize)
	if !assert.NoError(t, err, "expected no error reading") {
		return
	}

	if !assert.Equal(t, responseConf.ResponseString, response, "expected the configured response") {
		return
	}

	message := <-s.MessageChannel()
	if !assert.Equal(t, payload, message.Message, "expected same value") {
		return
	}

	assert.Len(t, s.GetErrors(), 0, "expected no errors")

}
