package server

import (
	"encoding/json"
	"fmt"
	"ws-chat/internal/chat"
	m "ws-chat/internal/model"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn     *websocket.Conn
	username string
	rooms    roomSet
	server   *Server
}

func NewClient(conn *websocket.Conn, server *Server) *Client {
	return &Client{
		conn:   conn,
		rooms:  make(roomSet),
		server: server,
	}
}

func (c *Client) Authenticated() bool {
	return c.username != ""
}

func (c *Client) IsInRoom(roomName string) bool {
	return c.rooms.Has(roomName)
}

func (c *Client) SendMessage(roomName, senderUsername, content string, timestamp int64) error {
	message := m.NewMessageResponse(roomName, senderUsername, content, timestamp)
	return c.conn.WriteJSON(message)
}

func (c *Client) sendError(message string) error {
	response := m.NewErrorResponse(message)
	return c.conn.WriteJSON(response)
}

func (c *Client) sendSuccess(message string) error {
	response := m.NewSuccessResponse(message)
	return c.conn.WriteJSON(response)
}

func (c *Client) sendHistory(roomName string, messages []chat.Message) error {
	historyMessages := make([]m.HistoryMessage, len(messages))
	for i, msg := range messages {
		historyMessages[i] = m.HistoryMessage{
			Username:  msg.Sender,
			Content:   msg.Content,
			Timestamp: msg.Timestamp,
		}
	}
	response := m.NewHistoryResponse(roomName, historyMessages)
	return c.conn.WriteJSON(response)
}

func (c *Client) handleCommand(data []byte) error {
	cmd, err := c.server.parseWsCommand(data)
	if err != nil {
		return err
	}

	handler, ok := commandHandlers[cmd.Command]
	if !ok {
		return fmt.Errorf("unknown command: %d", cmd.Command)
	}
	return handler(c, cmd.Payload)
}

var commandHandlers = map[m.Command]func(*Client, json.RawMessage) error{
	m.CmdSendMessage: (*Client).handleSendMessage,
	m.CmdJoinRoom:    (*Client).handleJoinRoom,
	m.CmdLeaveRoom:   (*Client).handleLeaveRoom,
}

func (c *Client) handleSendMessage(payload json.RawMessage) error {
	var msg m.SendMessagePayload
	if err := json.Unmarshal(payload, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal message payload: %w", err)
	}

	if !c.Authenticated() {
		return fmt.Errorf("user not authenticated")
	}

	if !c.IsInRoom(msg.RoomName) {
		return fmt.Errorf("not in room: %s", msg.RoomName)
	}

	timestamp, err := c.server.service.SendMessage(msg.RoomName, c.username, msg.Content)
	if err != nil {
		return err
	}

	c.server.broadcastToRoom(msg.RoomName, c.username, msg.Content, timestamp)
	return nil
}

func (c *Client) handleJoinRoom(payload json.RawMessage) error {
	var msg m.JoinRoomPayload
	if err := json.Unmarshal(payload, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal join room payload: %w", err)
	}

	if c.username == "" {
		c.username = c.server.usernameGenerator.Next()
	}

	if err := c.server.service.JoinRoom(msg.RoomName, c.username); err != nil {
		return err
	}

	c.rooms.Add(msg.RoomName)

	messages, err := c.server.service.GetMessages(msg.RoomName)
	if err != nil {
		return err
	}

	if err := c.sendHistory(msg.RoomName, messages); err != nil {
		return err
	}

	return c.sendSuccess(fmt.Sprintf("joined room: %s", msg.RoomName))
}

func (c *Client) handleLeaveRoom(payload json.RawMessage) error {
	var msg m.LeaveRoomPayload
	if err := json.Unmarshal(payload, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal leave room payload: %w", err)
	}

	if c.username == "" {
		return fmt.Errorf("user not authenticated")
	}

	if err := c.server.service.LeaveRoom(msg.RoomName, c.username); err != nil {
		return err
	}

	c.rooms.Remove(msg.RoomName)
	return c.sendSuccess(fmt.Sprintf("left room: %s", msg.RoomName))
}
