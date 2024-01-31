package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

type Message struct {
	from    net.Addr
	payload []byte
}

type Server struct {
	listenAddr string
	ln         net.Listener
	clientMap  map[*net.Addr]net.Conn
	sync.RWMutex

	msgch  chan Message
	quitch chan struct{}
}

func NewServer() *Server {
	return &Server{
		listenAddr: "localhost:5000",
		msgch:      make(chan Message, 10),           // max 10 message in queue
		clientMap:  make(map[*net.Addr]net.Conn, 10), // default support 10 client
		RWMutex:    sync.RWMutex{},
	}
}

func (s *Server) Start() error {

	ln, err := net.Listen("tcp", s.listenAddr)

	if err != nil {
		return err
	}

	log.Printf("listening server on %v ", ln.Addr().String())

	defer ln.Close()
	s.ln = ln

	go s.acceptLoop()
	go s.broadcastLoop()

	<-s.quitch

	return nil

}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.ln.Accept()

		if err != nil {
			log.Printf("error accepting connection, %v", err)
			continue
		}

		clientAddr := conn.RemoteAddr()

		go func(*Server) {
			s.RWMutex.Lock()
			s.clientMap[&clientAddr] = conn
			s.RWMutex.Unlock()
		}(s)

		go func(net.Addr) {
			notification := fmt.Sprintf("connection accepted from %v \n", clientAddr.String())
			s.broadcastMessage(clientAddr, []byte(notification))
		}(clientAddr)

		go s.readLoop(conn)
	}
}

func (s *Server) readLoop(conn net.Conn) {

	clientAddr := conn.RemoteAddr()

	defer func(net.Conn) {
		go conn.Close()

		go func(*Server) {
			s.RWMutex.Lock()
			delete(s.clientMap, &clientAddr)
			s.RWMutex.Unlock()
		}(s)

		go func(net.Addr) {
			notification := fmt.Sprintf("connection lost from %v \n", clientAddr)
			s.broadcastMessage(clientAddr, []byte(notification))
		}(clientAddr)

	}(conn)

	buf := make([]byte, 2048)

	for {
		n, err := conn.Read(buf)

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Printf("error reading data from connection, %v", err)
			continue
		}

		buf := buf[:n]

		go func([]byte) {
			s.broadcastMessage(clientAddr, buf)
		}(buf)

	}
}

func (s *Server) broadcastLoop() {

	for {
		msg := <-s.msgch

		for _, client := range s.clientMap {

			if client.RemoteAddr() == msg.from {
				continue
			}

			client := client

			go func(net.Conn) {
				client.Write(msg.payload)
			}(client)
		}
	}
}

func (s *Server) broadcastMessage(clientAddr net.Addr, payload []byte) {

	msg := Message{
		from:    clientAddr,
		payload: []byte(payload),
	}

	s.msgch <- msg
}

func main() {

	server := NewServer()

	log.Fatalf("error starting server %v", server.Start())
}
