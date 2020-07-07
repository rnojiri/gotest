package tcpudp

import (
	"bytes"
	"fmt"
	"net"
	"time"
)

//
// Functions to interact with the tcp server.
// author: rnojiri
//

const milisBetweenWrites time.Duration = 10

// ConnectTCP - connects to the specified address
func ConnectTCP(host string, port int, deadline time.Duration) (*net.TCPConn, error) {

	address, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, err
	}

	connection, err := net.DialTCP("tcp", nil, address)
	if err != nil {
		return nil, err
	}

	err = connection.SetDeadline(time.Now().Add(deadline))
	if err != nil {
		if connection != nil {
			connection.Close()
		}
		return nil, err
	}

	return connection, nil
}

// WriteTCP - writes to the connection
func WriteTCP(connection *net.TCPConn, payload string, endAfter bool) error {

	_, err := connection.Write(([]byte)(payload))
	if err != nil {
		return err
	}

	<-time.After(milisBetweenWrites * time.Millisecond)

	if endAfter {
		connection.Close()
	}

	return nil
}

// ReadTCP - read from the connection
func ReadTCP(connection *net.TCPConn, bufferSize int) (string, error) {

	readBuffer := make([]byte, bufferSize)

	_, err := connection.Read(readBuffer)
	if err != nil {
		return "", err
	}

	return string(bytes.Trim(readBuffer, "\x00")), nil
}
