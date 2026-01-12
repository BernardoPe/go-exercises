package server

import (
	"log"
	"sync"
	"sync/atomic"
	"ws-chat/internal/chat"

	"github.com/gorilla/websocket"
)

type Server struct {
	service           chat.Service
	clients           Clients
	mu                sync.RWMutex
	usernameGenerator *usernameGenerator
	closed            atomic.Bool
}

func NewServer(service chat.Service) *Server {
	return &Server{
		service:           service,
		clients:           make(map[*Client]struct{}),
		usernameGenerator: &usernameGenerator{},
	}
}

func (s *Server) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed.Load() {
		return nil
	}
	s.closed.Store(true)

	for client := range s.clients {
		client.conn.Close()
	}

	return nil
}

func (s *Server) addClient(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients.Add(client)
}

func (s *Server) removeClient(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients.Remove(client)
}

func (s *Server) handleClient(conn *websocket.Conn) {
	if s.closed.Load() {
		conn.Close()
		return
	}

	defer conn.Close()
	client := NewClient(conn, s)

	s.addClient(client)
	defer s.removeClient(client)
	log.Printf("new client connected: %v", conn.RemoteAddr())

	for {
		if err := s.handleIncoming(conn, client); err != nil {
			s.logReadError(err)
			break
		}
	}
}

func (s *Server) handleIncoming(conn *websocket.Conn, client *Client) error {
	_, message, err := conn.ReadMessage()
	if err != nil {
		return err
	}
	log.Printf("Received message from %s: %s", conn.RemoteAddr(), string(message))

	if err := client.handleCommand(message); err != nil {
		log.Printf("[ERROR] Client %s (user: %s): %v", conn.RemoteAddr(), client.username, err)
		if sendErr := client.sendError(err.Error()); sendErr != nil {
			log.Printf("[ERROR] Failed to send error response to %s: %v", conn.RemoteAddr(), sendErr)
		}
	}

	return nil
}

func (s *Server) logReadError(err error) {
	if err != nil && websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
		log.Printf("[ERROR] Unexpected WebSocket close: %v", err)
	}
}

func (s *Server) broadcastToRoom(roomName, username, content string, timestamp int64) {
	s.mu.RLock()
	s.clients.ForEach(func(client *Client) {
		if client.IsInRoom(roomName) {
			s.sendTo(client, roomName, username, content, timestamp)
		}
	})
	s.mu.RUnlock()
}

func (s *Server) sendTo(client *Client, roomName, username, content string, timestamp int64) {
	if !client.IsInRoom(roomName) {
		return
	}
	if err := client.SendMessage(roomName, username, content, timestamp); err != nil {
		log.Printf("[ERROR] Failed to send message to client %s in room %s: %v", username, roomName, err)
	}
}
