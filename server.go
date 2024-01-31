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
	quitch     chan struct{}
}

func NewServer() *Server {
	return &Server{
		listenAddr: "localhost:5000",
		msgch:      make(chan Message, 10),
		clientMap:  make(map[*net.Addr]net.Conn, 10),
	}
}

func (s *Server) Start() error {

	ln, err := net.Listen("tcp", s.listenAddr)

	if err != nil {
		return err
	}

	defer ln.Close()

	s.ln = ln

	<-s.quitch

	return nil

}
