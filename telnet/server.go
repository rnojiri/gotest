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
	readTimeout    time.Duration
	readBufferSize int
	port           int
	started        bool
}

// MessageData - the message data received
type MessageData struct {
	Message string
	Date    time.Time
}

// NewServer - creates a new telnet server on a random port
func NewServer(host string, messageChannelSize int, readBufferSize int, readTimeout time.Duration, start bool) (*Server, int) {

	var listener net.Listener
	var port int
	var err error

	for i := 0; i < listenRetries; i++ {

		port = utils.GeneratePort()

		listener, err = net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
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
		messageChannel: make(chan MessageData, messageChannelSize),
		readTimeout:    readTimeout,
		listener:       listener,
		readBufferSize: readBufferSize,
		port:           port,
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

		err = conn.SetReadDeadline(time.Now().Add(ts.readTimeout))
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
		readBuffer := make([]byte, ts.readBufferSize)
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
