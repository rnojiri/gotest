package tcpudp

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"time"

	utils "github.com/uol/gotest/utils"
)

//
// Creates a test tcp server.
// author: rnojiri
//

// TCPServer - the tcp server
type TCPServer struct {
	listener      net.Listener
	configuration *TCPConfiguration
	server
}

// TCPConfiguration - the tcp server configuration
type TCPConfiguration struct {
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	ResponseString string
	ServerConfiguration
}

// NewTCPServer - creates a new telnet server on a random port
func NewTCPServer(configuration *TCPConfiguration, start bool) (*TCPServer, int) {

	var listener net.Listener
	var port int
	var err error

	for i := 0; i < listenRetries; i++ {

		port = utils.GeneratePort()
		address, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", configuration.Host, port))
		if err != nil {
			panic(err)
		}

		listener, err = net.ListenTCP("tcp", address)
		if err != nil {
			if strings.Contains(err.Error(), "address already in use") {
				<-time.After(time.Second)
				fmt.Println("port already in use, trying another...")
			} else {
				panic(err)
			}
		} else {
			break
		}
	}

	if err != nil {
		panic(err)
	}

	server := &TCPServer{
		server: server{
			messageChannel: make(chan MessageData, configuration.MessageChannelSize),
			port:           port,
		},
		listener:      listener,
		configuration: configuration,
	}

	if start {
		server.Start()
	}

	return server, port
}

// Start - starts the server to receive connections
func (ts *TCPServer) Start() {

	if ts.started {
		return
	}

	ts.started = true

	go ts.startListeningLoop()
}

func (ts *TCPServer) startListeningLoop() {

	for {
		conn, err := ts.listener.Accept()
		if err != nil {
			ts.errors = append(ts.errors, err)
			return
		}

		ts.handleConnection(conn)
	}
}

// Stop - stops the server
func (ts *TCPServer) Stop() error {

	return ts.listener.Close()
}

// handleConnection - handles the current connection
func (ts *TCPServer) handleConnection(conn net.Conn) {

	buffer := bytes.Buffer{}

	for {
		err := conn.SetReadDeadline(time.Now().Add(ts.configuration.ReadTimeout))
		if err != nil {
			ts.errors = append(ts.errors, err)
			return
		}

		readBuffer := make([]byte, ts.configuration.ReadBufferSize)
		n, err := conn.Read(readBuffer)
		if err != nil {
			if nErr, ok := err.(net.Error); ok {
				if nErr.Timeout() {
					break
				}
			}

			ts.errors = append(ts.errors, err)
			return
		}

		if n == 0 {
			break
		}

		_, err = buffer.Write(bytes.Trim(readBuffer, "\x00"))
		if err != nil {
			return
		}
	}

	err := conn.SetWriteDeadline(time.Now().Add(ts.configuration.WriteTimeout))
	if err != nil {
		ts.errors = append(ts.errors, err)
		return
	}

	if len(ts.configuration.ResponseString) > 0 {
		_, err := conn.Write(([]byte)(ts.configuration.ResponseString))
		if err != nil {
			ts.errors = append(ts.errors, err)
			return
		}
	}

	ts.messageChannel <- MessageData{
		Message: buffer.String(),
		Date:    time.Now(),
	}
}

// MessageChannel - reads from the message channel
func (ts *TCPServer) MessageChannel() <-chan MessageData {

	return ts.messageChannel
}

// GetErrors - get asynchronous errors
func (ts *TCPServer) GetErrors() []error {

	return ts.errors
}
