package telnet

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"time"

	utils "github.com/uol/gotest/utils"
)

const listenRetries int = 10

// Server - the telnet server
type Server struct {
	listener       net.Listener
	errors         []error
	messageChannel chan MessageData
	port           int
	started        bool
	configuration  *Configuration
}

// MessageData - the message data received
type MessageData struct {
	Message string
	Date    time.Time
}

// Configuration - the server configuration
type Configuration struct {
	Host               string
	MessageChannelSize int
	ReadBufferSize     int
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	ResponseString     string
}

// NewServer - creates a new telnet server on a random port
func NewServer(configuration *Configuration, start bool) (*Server, int) {

	var listener net.Listener
	var port int
	var err error

	for i := 0; i < listenRetries; i++ {

		port = utils.GeneratePort()

		listener, err = net.Listen("tcp", fmt.Sprintf("%s:%d", configuration.Host, port))
		if err != nil {
			if strings.Contains(err.Error(), "address already in use") {
				<-time.After(time.Second)
				fmt.Println("port already in use, trying another...")
			}
		} else {
			break
		}
	}

	if err != nil {
		panic(err)
	}

	server := &Server{
		messageChannel: make(chan MessageData, configuration.MessageChannelSize),
		listener:       listener,
		port:           port,
		configuration:  configuration,
	}

	if start {
		server.Start()
	}

	return server, port
}

// Start - starts the server to receive connections
func (ts *Server) Start() {

	if ts.started {
		return
	}

	ts.started = true

	go ts.startListeningLoop()
}

func (ts *Server) startListeningLoop() {

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
func (ts *Server) Stop() error {

	close(ts.messageChannel)

	return ts.listener.Close()
}

// handleConnection - handles the current connection
func (ts *Server) handleConnection(conn net.Conn) {

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
func (ts *Server) MessageChannel() <-chan MessageData {

	return ts.messageChannel
}

// GetErrors - get asynchronous errors
func (ts *Server) GetErrors() []error {

	return ts.errors
}
