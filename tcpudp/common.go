package tcpudp

import "time"

//
// Commons artifacts for the servers.
// author: rnojiri
//

const listenRetries int = 10

// MessageData - the message data received
type MessageData struct {
	Message string
	Date    time.Time
	Host    string
	Port    int
}

// ServerConfiguration - common configuration
type ServerConfiguration struct {
	MessageTimeout     time.Duration
	Host               string
	MessageChannelSize int
	ReadBufferSize     int
}

func (sc *ServerConfiguration) setDefaults() {

	if sc.ReadBufferSize <= 0 {
		sc.ReadBufferSize = 4096
	}

	if sc.MessageChannelSize <= 0 {
		sc.MessageChannelSize = 100
	}

	if sc.MessageTimeout <= 0 {
		sc.MessageTimeout = 5 * time.Second
	}

	if sc.Host == "" {
		sc.Host = "localhost"
	}
}

// server - core
type server struct {
	errors         []error
	messageChannel chan MessageData
	port           int
	started        bool
}
