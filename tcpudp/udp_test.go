package tcpudp_test

import (
	"fmt"
	"testing"
	"time"

	tcpudp "github.com/rnojiri/gotest/tcpudp"
	gotest "github.com/rnojiri/gotest/utils"
	"github.com/stretchr/testify/assert"
)

//
// Tests for the udp server.
// author: rnojiri
//

var (
	defaultUDPConf tcpudp.ServerConfiguration = tcpudp.ServerConfiguration{
		Host:               testHost,
		MessageChannelSize: numMsgChan,
		ReadBufferSize:     bufferSize,
	}
)

// TestUDPCreateServer - tests creating the server only (not accepting connections)
func TestUDPCreateServer(t *testing.T) {

	s, p := tcpudp.NewUDPServer(&defaultUDPConf, false)

	if !assert.NotNil(t, s, "expected a valid instance") {
		return
	}

	assert.GreaterOrEqual(t, p, 10000, "expected port greater than 10000")
}

func mustCreateUDPServer(autoStart bool, readTimeout time.Duration) (*tcpudp.UDPServer, int) {

	return tcpudp.NewUDPServer(&defaultUDPConf, autoStart)
}

// TestUDPNotStartedServerStop - tests the stop server function
func TestUDPNotStartedServerStop(t *testing.T) {

	s, _ := mustCreateUDPServer(false, time.Second)
	if s == nil {
		return
	}

	err := s.Stop()
	assert.NoError(t, err, "error not expected")
}

// TestUDPStartedServerStop - tests the stop server function
func TestUDPStartedServerStop(t *testing.T) {

	s, _ := mustCreateUDPServer(true, time.Second)
	if s == nil {
		return
	}

	err := s.Stop()
	assert.NoError(t, err, "error not expected")
}

// TestUDPOneMessage - tests the server with only one message
func TestUDPOneMessage(t *testing.T) {

	s, port := mustCreateUDPServer(true, 5*time.Second)
	defer s.Stop()
	if s == nil {
		return
	}

	conn, err := tcpudp.ConnectUDP(testHost, port, time.Second)
	if !assert.NoError(t, err, "expected no error connecting") {
		return
	}

	payload := "test"
	now := time.Now()

	err = tcpudp.WriteUDP(conn, payload)
	if !assert.NoError(t, err, "expected no error writing") {
		return
	}

	message := <-s.MessageChannel()
	if !testMessage(t, &message, payload, port, now) {
		return
	}

	assert.Len(t, s.GetErrors(), 0, "expected no errors")
}

// TestUDPMultipleMessages - tests the server with multiple messages
func TestUDPMultipleMessages(t *testing.T) {

	s, port := mustCreateUDPServer(true, time.Second)
	defer s.Stop()
	if s == nil {
		return
	}

	conn, err := tcpudp.ConnectUDP(testHost, port, time.Second)
	if !assert.NoError(t, err, "expected no error connecting") {
		return
	}

	messageFormat := "test%d"
	numMessages := gotest.RandomInt(2, 10)
	payloads := make([]string, numMessages)
	times := make([]time.Time, numMessages)

	for i := 0; i < numMessages; i++ {

		payloads[i] = fmt.Sprintf(messageFormat, i)
		times[i] = time.Now()

		err = tcpudp.WriteUDP(conn, payloads[i])
		if !assert.NoError(t, err, "expected no error writing") {
			return
		}
	}

	for i := 0; i < numMessages; i++ {

		message := <-s.MessageChannel()
		if !testMessage(t, &message, payloads[i], port, times[i]) {
			return
		}
	}

	assert.Len(t, s.GetErrors(), 0, "expected no errors")
}
