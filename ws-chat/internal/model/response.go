package model

const (
	ResponseTypeMessage = "message"
	ResponseTypeError   = "error"
	ResponseTypeSuccess = "success"
	ResponseTypeHistory = "history"
)

type MessageResponse struct {
	Type      string `json:"type"`
	Room      string `json:"room"`
	Username  string `json:"username"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
}

func NewMessageResponse(room, username, content string, timestamp int64) MessageResponse {
	return MessageResponse{
		Type:      ResponseTypeMessage,
		Room:      room,
		Username:  username,
		Content:   content,
		Timestamp: timestamp,
	}
}

type HistoryMessage struct {
	Username  string `json:"username"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
}

type HistoryResponse struct {
	Type     string           `json:"type"`
	Room     string           `json:"room"`
	Messages []HistoryMessage `json:"messages"`
}

func NewHistoryResponse(room string, messages []HistoryMessage) HistoryResponse {
	return HistoryResponse{
		Type:     ResponseTypeHistory,
		Room:     room,
		Messages: messages,
	}
}

type ErrorResponse struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func NewErrorResponse(message string) ErrorResponse {
	return ErrorResponse{
		Type:    ResponseTypeError,
		Message: message,
	}
}

type SuccessResponse struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func NewSuccessResponse(message string) SuccessResponse {
	return SuccessResponse{
		Type:    ResponseTypeSuccess,
		Message: message,
	}
}
