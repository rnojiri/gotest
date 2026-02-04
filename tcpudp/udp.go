package tcpudp

import (
	"fmt"
	"net"
	"strings"
	"time"

	utils "github.com/rnojiri/gotest/utils"
)

//
// Creates a test udp server.
// author: rnojiri
//

// UDPServer - the udp server
type UDPServer struct {
	listener      *net.UDPConn
	configuration *ServerConfiguration
	server
}

// NewUDPServer - creates a new udp server on a random port
func NewUDPServer(configuration *ServerConfiguration, start bool) (*UDPServer, int) {

	if configuration == nil {
		configuration = &ServerConfiguration{}
	}

	configuration.setDefaults()

	var listener *net.UDPConn
	var port int

	for i := 0; i < listenRetries; i++ {

		port = utils.GeneratePort()
		address, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", configuration.Host, port))
		if err != nil {
			panic(err)
		}

		listener, err = net.ListenUDP("udp", address)
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

	server := &UDPServer{
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
func (us *UDPServer) Start() {

	if us.started {
		return
	}

	us.started = true

	go us.startListeningLoop()
}

func (us *UDPServer) startListeningLoop() {

	for {
		buffer := make([]byte, us.configuration.ReadBufferSize)

		rlen, err := us.listener.Read(buffer)
		if err != nil {
			// check if connection was closed
			if strings.Contains(err.Error(), "use of closed network connection") {
				return
			}

			panic(err)
		}

		us.handlePacket(buffer[0:rlen])
	}
}

// Stop - stops the server
func (us *UDPServer) Stop() error {

	return us.listener.Close()
}

// handlePacket - handles the current connection
func (us *UDPServer) handlePacket(buffer []byte) {

	us.messageChannel <- MessageData{
		Message: string(buffer),
		Date:    time.Now(),
		Host:    us.configuration.Host,
		Port:    us.port,
	}
}

// GetMessage - get the message from channel
func (us *UDPServer) GetMessage() *MessageData {

	timeout := time.After(us.configuration.MessageTimeout)

	for {
		select {
		case <-timeout:
			return nil
		case msg := <-us.messageChannel:
			return &msg
		default:
			<-time.After(100 * time.Millisecond)
		}
	}
}

// GetErrors - get asynchronous errors
func (us *UDPServer) GetErrors() []error {

	return us.errors
}
