package chat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServiceSendMessage(t *testing.T) {
	repo := NewInMemoryRoomRepository()
	service := NewService(repo)

	timestamp, err := service.SendMessage("general", "alice", "hello world")
	assert.NoError(t, err)
	assert.Greater(t, timestamp, int64(0))

	room, exists := repo.Get("general")
	assert.True(t, exists)
	assert.Equal(t, 1, len(room.Messages))
	assert.Equal(t, "alice", room.Messages[0].Sender)
	assert.Equal(t, "hello world", room.Messages[0].Content)

	timestamp2, err := service.SendMessage("general", "bob", "hi alice")
	assert.NoError(t, err)
	assert.Greater(t, timestamp2, int64(0))

	room, _ = repo.Get("general")
	assert.Equal(t, 2, len(room.Messages))
}

func TestServiceJoinRoom(t *testing.T) {
	repo := NewInMemoryRoomRepository()
	service := NewService(repo)

	// Joining non-existent room should auto-create it
	err := service.JoinRoom("general", "alice")
	assert.NoError(t, err)

	retrieved, exists := repo.Get("general")
	assert.True(t, exists)
	assert.True(t, retrieved.hasUser("alice"))

	// Joining same room again should error
	err = service.JoinRoom("general", "alice")
	assert.Error(t, err)

	// Joining new room should succeed
	err = service.JoinRoom("random", "bob")
	assert.NoError(t, err)
}

func TestServiceLeaveRoom(t *testing.T) {
	repo := NewInMemoryRoomRepository()
	service := NewService(repo)

	room := &Room{Name: "general", Users: RoomUsers{"alice", "bob"}}
	err := repo.Save(room)
	assert.NoError(t, err)

	err = service.LeaveRoom("general", "alice")
	assert.NoError(t, err)

	retrieved, _ := repo.Get("general")
	assert.False(t, retrieved.hasUser("alice"))
	assert.True(t, retrieved.hasUser("bob"))

	err = service.LeaveRoom("general", "charlie")
	assert.Error(t, err)

	err = service.LeaveRoom("nonexistent", "alice")
	assert.Error(t, err)
}
