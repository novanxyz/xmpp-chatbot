package main

import (
	"net"
	"fmt"
)

func mainx() {
	conn, err := net.Dial("tcp", ":5222")
	if err != nil {
		fmt.Errorf("Couldn't connect")
	}
	message := "<?xml version='1.0'?>" +
          "<stream:stream to='localhost' version='1.0' " +
          "xmlns='jabber:client' xmlns:stream='http://etherx.jabber.org/streams'>"
	conn.Write([]byte(message))

	recvBuf := make([]byte, 4096)
    conn.Read(recvBuf)
    fmt.Print("Message Received:", string(recvBuf))
}