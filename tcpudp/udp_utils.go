package tcpudp

import (
	"fmt"
	"net"
	"time"
)

//
// Functions to interact with the udp server.
// author: rnojiri
//

// ConnectUDP - connects to the specified address
func ConnectUDP(host string, port int, deadline time.Duration) (*net.UDPConn, error) {

	address, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, err
	}

	connection, err := net.DialUDP("udp", nil, address)
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

// WriteUDP - writes to the connection
func WriteUDP(connection *net.UDPConn, payload string) error {

	_, err := fmt.Fprint(connection, payload)
	if err != nil {
		return err
	}

	<-time.After(milisBetweenWrites * time.Millisecond)

	return nil
}
