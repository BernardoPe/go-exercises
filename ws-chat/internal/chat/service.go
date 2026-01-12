package chat

import (
	"time"
)

type Service interface {
	SendMessage(roomName, username, content string) (int64, error)
	JoinRoom(roomName, username string) error
	LeaveRoom(roomName, username string) error
	GetMessages(roomName string) ([]Message, error)
}

type chatService struct {
	repo RoomRepository
}

func NewService(repo RoomRepository) Service {
	return &chatService{
		repo: repo,
	}
}

func (s *chatService) SendMessage(roomName, username, content string) (int64, error) {
	message := Message{
		Sender:    username,
		Content:   content,
		Timestamp: time.Now().Unix(),
	}
	if err := s.repo.Add(roomName, message); err != nil {
		return 0, err
	}
	return message.Timestamp, nil
}

func (s *chatService) JoinRoom(roomName, username string) error {
	return s.repo.AddUser(roomName, username)
}

func (s *chatService) LeaveRoom(roomName, username string) error {
	return s.repo.RemoveUser(roomName, username)
}

func (s *chatService) GetMessages(roomName string) ([]Message, error) {
	return s.repo.GetMessages(roomName)
}
