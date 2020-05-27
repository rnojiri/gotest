package telnet_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	gotesttelnet "github.com/uol/gotest/telnet"
	gotest "github.com/uol/gotest/utils"
)

const (
	testHost   string = "localhost"
	numMsgChan int    = 100
	bufferSize int    = 256
)

// TestCreateServer - tests creating the server only (not accepting connections)
func TestCreateServer(t *testing.T) {

	s, p := gotesttelnet.NewServer(testHost, numMsgChan, bufferSize, time.Second, false)

	if !assert.NotNil(t, s, "expected a valid instance") {
		return
	}

	assert.GreaterOrEqual(t, p, 10000, "expected port greater than 10000")
}

func mustCreateServer(autoStart bool, readTimeout time.Duration) (*gotesttelnet.Server, int) {

	s, port := gotesttelnet.NewServer(testHost, numMsgChan, bufferSize, readTimeout, autoStart)

	return s, port
}

// TestNotStartedServerStop - tests the stop server function
func TestNotStartedServerStop(t *testing.T) {

	s, _ := mustCreateServer(false, time.Second)
	if s == nil {
		return
	}

	err := s.Stop()
	assert.NoError(t, err, "error not expected")
}

// TestStartedServerStop - tests the stop server function
func TestStartedServerStop(t *testing.T) {

	s, _ := mustCreateServer(true, time.Second)
	if s == nil {
		return
	}

	err := s.Stop()
	assert.NoError(t, err, "error not expected")
}

// TestOneMessage - tests the server with only one message
func TestOneMessage(t *testing.T) {

	s, port := mustCreateServer(true, time.Second)
	defer s.Stop()
	if s == nil {
		return
	}

	conn, err := gotesttelnet.Connect(testHost, port, time.Second)
	if !assert.NoError(t, err, "expected no error connecting") {
		return
	}

	payload := "test"

	err = gotesttelnet.Write(conn, payload)
	if !assert.NoError(t, err, "expected no error writing") {
		return
	}

	message := <-s.MessageChannel()
	if !assert.Equal(t, payload, message.Message, "expected same value") {
		return
	}

	assert.Len(t, s.GetErrors(), 0, "expected no errors")
}

// TestMultipleMessages - tests the server with multiple messages
func TestMultipleMessages(t *testing.T) {

	s, port := mustCreateServer(true, time.Second)
	defer s.Stop()
	if s == nil {
		return
	}

	conn, err := gotesttelnet.Connect(testHost, port, time.Second)
	if !assert.NoError(t, err, "expected no error connecting") {
		return
	}

	messageFormat := "test%d\n"
	numMessages := gotest.RandomInt(2, 10)
	for i := 0; i < numMessages; i++ {

		payload := fmt.Sprintf(messageFormat, i)

		err = gotesttelnet.Write(conn, payload)
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
