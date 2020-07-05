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
}

// ServerConfiguration - common configuration
type ServerConfiguration struct {
	Host               string
	MessageChannelSize int
	ReadBufferSize     int
}

// server - core
type server struct {
	errors         []error
	messageChannel chan MessageData
	port           int
	started        bool
}
