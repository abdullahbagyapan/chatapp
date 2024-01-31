package main

import (
	"net"
)

type Message struct {
	from    string
	payload []byte
}

type Server struct {
	listenAddr string
	ln         net.Listener
	clientMap  map[*net.Addr]net.Conn
	msgch      chan Message
}
