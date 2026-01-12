package client

import (
	"encoding/json"
	"fmt"
	"strings"
	"ws-chat/internal/model"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// UI Mostly AI generated

var (
	messageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	usernameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true)

	timestampStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	roomStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("99")).
			Bold(true)

	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("120"))
)

type TUIModel struct {
	client      *Client
	messages    []string
	input       string
	currentRoom string
	ready       bool
	width       int
	height      int
	maxMessages int
	err         error
}

type serverMessage struct {
	data []byte
}

func NewTUIModel(client *Client) TUIModel {
	return TUIModel{
		client:      client,
		messages:    []string{},
		maxMessages: 100,
		ready:       false,
	}
}

func (m TUIModel) Init() tea.Cmd {
	return tea.Batch(
		waitForMessages(m.client),
		tea.EnterAltScreen,
	)
}

func (m TUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyEnter:
			if m.input == "" {
				return m, nil
			}

			input := m.input
			m.input = ""

			// Check if it's a /room command to update UI state
			if strings.HasPrefix(input, "/room ") {
				parts := strings.Fields(input)
				if len(parts) >= 2 {
					m.currentRoom = parts[1]
				}
			}

			if err := m.client.HandleInput(input); err != nil {
				if err == ErrQuit {
					return m, tea.Quit
				}
				m.addMessage(errorStyle.Render(fmt.Sprintf("Error: %v", err)))
			}
			return m, nil

		case tea.KeyBackspace:
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
			return m, nil

		case tea.KeySpace:
			m.input += " "
			return m, nil

		case tea.KeyRunes:
			m.input += string(msg.Runes)
			return m, nil
		default:
			return m, nil
		}

	case serverMessage:
		m.processServerMessage(msg.data)
		return m, waitForMessages(m.client)

	case error:
		m.err = msg
		return m, nil
	}

	return m, nil
}

func (m TUIModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var b strings.Builder

	// Calculate available height for messages (reserve space for input, borders, and help)
	messagesHeight := m.height - 2 // Reserve space for separator, input, help, and padding

	visibleMessages := m.messages
	if len(visibleMessages) > messagesHeight {
		visibleMessages = visibleMessages[len(visibleMessages)-messagesHeight:]
	}

	for _, msg := range visibleMessages {
		b.WriteString(msg)
		b.WriteString("\n")
	}

	// Fill remaining space
	for i := len(visibleMessages); i < messagesHeight; i++ {
		b.WriteString("\n")
	}

	// Render separator
	b.WriteString(strings.Repeat("─", m.width))
	b.WriteString("\n")

	// Render input area
	prompt := "> "
	if m.currentRoom != "" {
		prompt = roomStyle.Render(fmt.Sprintf("[%s]", m.currentRoom)) + " > "
	}

	b.WriteString(prompt)
	b.WriteString(inputStyle.Render(m.input))
	b.WriteString(lipgloss.NewStyle().Render("│")) // Cursor placeholder
	b.WriteString("\n")

	// Render help text
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
	helpText := "Commands: /join <room> | /room <room> | /leave | /quit or Ctrl+C | Esc to exit"
	b.WriteString(helpStyle.Render(helpText))

	return b.String()
}

func (m *TUIModel) addMessage(msg string) {
	m.messages = append(m.messages, msg)
	if len(m.messages) > m.maxMessages {
		m.messages = m.messages[len(m.messages)-m.maxMessages:]
	}
}

func (m *TUIModel) processServerMessage(data []byte) {
	var typeCheck struct {
		Type string `json:"type"`
	}

	if err := json.Unmarshal(data, &typeCheck); err != nil {
		m.addMessage(errorStyle.Render(fmt.Sprintf("Failed to parse message: %v", err)))
		return
	}

	switch typeCheck.Type {
	case model.ResponseTypeMessage:
		var resp model.MessageResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			m.addMessage(errorStyle.Render(fmt.Sprintf("Failed to parse message response: %v", err)))
			return
		}
		timeStr := formatTimestamp(resp.Timestamp)
		msg := fmt.Sprintf("%s %s %s: %s",
			roomStyle.Render(fmt.Sprintf("[%s]", resp.Room)),
			timestampStyle.Render(timeStr),
			usernameStyle.Render(resp.Username),
			messageStyle.Render(resp.Content))
		m.addMessage(msg)

	case model.ResponseTypeError:
		var resp model.ErrorResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			m.addMessage(errorStyle.Render(fmt.Sprintf("Failed to parse error response: %v", err)))
			return
		}
		m.addMessage(errorStyle.Render(fmt.Sprintf("%s", resp.Message)))

	case model.ResponseTypeSuccess:
		var resp model.SuccessResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			m.addMessage(errorStyle.Render(fmt.Sprintf("Failed to parse success response: %v", err)))
			return
		}

		if strings.HasPrefix(resp.Message, "joined room:") {
			roomName := strings.TrimPrefix(resp.Message, "joined room: ")
			m.currentRoom = strings.TrimSpace(roomName)
			m.addMessage(successStyle.Render(resp.Message))
		} else if strings.HasPrefix(resp.Message, "left room:") {
			m.currentRoom = ""
			m.addMessage(successStyle.Render(resp.Message))
		}

	case model.ResponseTypeHistory:
		var resp model.HistoryResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			m.addMessage(errorStyle.Render(fmt.Sprintf("Failed to parse history response: %v", err)))
			return
		}
		if len(resp.Messages) == 0 {
			m.addMessage(messageStyle.Render("No messages yet"))
		} else {
			for _, msg := range resp.Messages {
				timeStr := formatTimestamp(msg.Timestamp)
				m.addMessage(fmt.Sprintf("[%s] %s %s: %s",
					roomStyle.Render(resp.Room),
					timestampStyle.Render(timeStr),
					usernameStyle.Render(msg.Username),
					messageStyle.Render(msg.Content)))
			}
		}
	}
}

func waitForMessages(client *Client) tea.Cmd {
	return func() tea.Msg {
		client.mu.RLock()
		conn := client.conn
		client.mu.RUnlock()

		if conn == nil {
			return fmt.Errorf("not connected")
		}

		_, data, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		return serverMessage{data: data}
	}
}
