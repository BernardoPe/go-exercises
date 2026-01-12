package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"ws-chat/internal/model"

	"github.com/gorilla/websocket"
)

var ErrQuit = errors.New("quit requested")

type Client struct {
	serverAddr  string
	currentRoom string
	conn        *websocket.Conn
	mu          sync.RWMutex
}

func NewClient(serverAddr string) (*Client, error) {
	if serverAddr == "" {
		return nil, errors.New("server address is required")
	}

	return &Client{
		serverAddr: serverAddr,
	}, nil
}

func (c *Client) Connect(ctx context.Context) error {
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, c.serverAddr, nil)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	return nil
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) ReadMessages(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			c.mu.RLock()
			conn := c.conn
			c.mu.RUnlock()

			if conn == nil {
				return
			}

			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					fmt.Printf("WebSocket error: %v", err)
				}
				return
			}
			c.displayMessage(message)
		}
	}
}

func (c *Client) HandleInput(input string) error {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}

	if strings.HasPrefix(input, "/") {
		return c.handleCommand(input)
	}

	c.mu.RLock()
	room := c.currentRoom
	c.mu.RUnlock()

	if room == "" {
		return errors.New("you must join a room first. Use /join <room>")
	}

	payload := model.SendMessagePayload{
		RoomName: room,
		Content:  input,
	}

	return c.sendCommand(model.CmdSendMessage, payload)
}

func (c *Client) handleCommand(cmd string) error {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return nil
	}
	switch parts[0] {
	case "/join":
		if len(parts) < 2 {
			return errors.New("usage: /join <room>")
		}
		return c.joinRoom(parts[1])
	case "/leave":
		return c.leaveRoom()
	case "/room":
		if len(parts) < 2 {
			return errors.New("usage: /room <room>")
		}
		return c.setCurrentRoom(parts[1])
	case "/quit", "/exit":
		return ErrQuit
	default:
		return fmt.Errorf("unknown command: %s", parts[0])
	}
}

func (c *Client) joinRoom(room string) error {
	payload := model.JoinRoomPayload{
		RoomName: room,
	}

	if err := c.sendCommand(model.CmdJoinRoom, payload); err != nil {
		return err
	}

	c.mu.Lock()
	c.currentRoom = room
	c.mu.Unlock()

	return nil
}

func (c *Client) leaveRoom() error {
	c.mu.RLock()
	room := c.currentRoom
	c.mu.RUnlock()

	if room == "" {
		return errors.New("you are not in any room")
	}

	payload := model.LeaveRoomPayload{
		RoomName: room,
	}

	if err := c.sendCommand(model.CmdLeaveRoom, payload); err != nil {
		return err
	}

	c.mu.Lock()
	c.currentRoom = ""
	c.mu.Unlock()

	return nil
}

func (c *Client) setCurrentRoom(room string) error {
	c.mu.Lock()
	c.currentRoom = room
	c.mu.Unlock()
	return nil
}

func (c *Client) sendCommand(cmd model.Command, payload interface{}) error {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return errors.New("not connected")
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	wsCmd := model.NewWsCommand(cmd, payloadBytes)
	return conn.WriteJSON(wsCmd)
}

func (c *Client) displayMessage(data []byte) {
	var typeCheck struct {
		Type string `json:"type"`
	}

	if err := json.Unmarshal(data, &typeCheck); err != nil {
		c.printWithPrompt("\nFailed to parse message: %v\n", err)
		return
	}

	switch typeCheck.Type {
	case model.ResponseTypeMessage:
		var resp model.MessageResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			c.printWithPrompt("\nFailed to parse message response: %v\n", err)
			return
		}
		timeStr := formatTimestamp(resp.Timestamp)
		c.printWithPrompt("\n[%s] %s %s: %s\n", resp.Room, timeStr, resp.Username, resp.Content)

	case model.ResponseTypeError:
		var resp model.ErrorResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			log.Printf("\nFailed to parse error response: %v\n", err)
			fmt.Printf("> ")
			return
		}
		c.printWithPrompt("\nError: %s\n", resp.Message)

	case model.ResponseTypeSuccess:
		var resp model.SuccessResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			c.printWithPrompt("\nFailed to parse success response: %v\n", err)
			return
		}
		c.printWithPrompt("\n%s\n", resp.Message)

	case model.ResponseTypeHistory:
		var resp model.HistoryResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			c.printWithPrompt("\nFailed to parse history response: %v\n", err)
			return
		}
		if len(resp.Messages) == 0 {
			fmt.Println("No messages yet")
		} else {
			for _, msg := range resp.Messages {
				timeStr := formatTimestamp(msg.Timestamp)
				fmt.Printf("[%s] %s %s: %s\n", resp.Room, timeStr, msg.Username, msg.Content)
			}
		}

	default:
		var prettyData map[string]interface{}
		if err := json.Unmarshal(data, &prettyData); err == nil {
			if formatted, err := json.MarshalIndent(prettyData, "", "  "); err == nil {
				c.printWithPrompt("%s\n", string(formatted))
			}
		}
	}
}

func formatTimestamp(timestamp int64) string {
	t := time.Unix(timestamp, 0)
	return t.Format("15:04:05")
}

func (c *Client) printWithPrompt(format string, args ...interface{}) {
	fmt.Printf(format, args...)
	fmt.Print("> ")
}
