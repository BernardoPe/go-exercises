package server

import (
	"fmt"
	"http_server/internal/request"
	response "http_server/internal/response"
	"log"
	"net"
	"sync/atomic"
)

type Server struct {
	Listener net.Listener
	Port     int
	closed   atomic.Bool
}

type Handler func(w *response.Writer, req *request.Request)

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	server := &Server{
		Listener: listener,
		Port:     port,
	}
	go server.listen(handler)
	return &Server{
		Listener: listener,
		Port:     port,
	}, nil
}

func (s *Server) listen(handler Handler) {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Println("Error accepting connection:", err)
			continue
		}
		log.Printf("Connection from %s\n", conn.RemoteAddr().String())
		go s.handle(conn, handler)
	}
}

func (s *Server) handle(conn net.Conn, handler Handler) {
	defer conn.Close()

	req, err := request.FromReader(conn)
	if err != nil {
		log.Println("Error reading request:", err)
		return
	}

	w := response.NewWriter(conn)
	handler(w, req)
}

func (s *Server) Close() error {
	if s.closed.Load() {
		return nil
	}
	s.closed.Store(true)
	return s.Listener.Close()
}

func (s *Server) Addr() net.Addr {
	return s.Listener.Addr()
}
