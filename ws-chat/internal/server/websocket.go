package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	m "ws-chat/internal/model"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func (s *Server) HandleWebSocket() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, "Failed to upgrade to WebSocket", http.StatusInternalServerError)
			return
		}
		s.handleClient(conn)
	}
}

func (s *Server) parseWsCommand(data []byte) (*m.WsCommand, error) {
	var cmd m.WsCommand
	if err := json.Unmarshal(data, &cmd); err != nil {
		return nil, fmt.Errorf("failed to parse command: %w", err)
	}
	return &cmd, nil
}
