package chat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInMemoryRoomRepository(t *testing.T) {
	repo := NewInMemoryRoomRepository()

	_, exists := repo.Get("general")
	assert.False(t, exists)

	room := &Room{Name: "general"}
	err := repo.Save(room)
	assert.NoError(t, err)

	retrieved, exists := repo.Get("general")
	assert.True(t, exists)
	assert.Equal(t, "general", retrieved.Name)
}

func TestInMemoryRoomRepositoryAdd(t *testing.T) {
	repo := NewInMemoryRoomRepository()

	msg := Message{Sender: "alice", Content: "hello", Timestamp: 1000}
	err := repo.Add("general", msg)
	assert.NoError(t, err)

	room, exists := repo.Get("general")
	assert.True(t, exists)
	assert.Equal(t, 1, len(room.Messages))
	assert.Equal(t, "hello", room.Messages[0].Content)

	msg2 := Message{Sender: "bob", Content: "world", Timestamp: 2000}
	err = repo.Add("general", msg2)
	assert.NoError(t, err)

	room, _ = repo.Get("general")
	assert.Equal(t, 2, len(room.Messages))
}

func TestInMemoryRoomRepositoryAddUser(t *testing.T) {
	repo := NewInMemoryRoomRepository()

	// Auto-creates room if it doesn't exist
	err := repo.AddUser("general", "alice")
	assert.NoError(t, err)

	room, exists := repo.Get("general")
	assert.True(t, exists)
	assert.True(t, room.hasUser("alice"))

	// Adding same user again should error
	err = repo.AddUser("general", "alice")
	assert.Error(t, err)

	retrieved, _ := repo.Get("general")
	assert.Equal(t, 1, len(retrieved.Users))
	assert.True(t, retrieved.hasUser("alice"))
}

func TestInMemoryRoomRepositoryRemoveUser(t *testing.T) {
	repo := NewInMemoryRoomRepository()

	err := repo.RemoveUser("general", "alice")
	assert.Error(t, err)

	room := &Room{Name: "general", Users: RoomUsers{"alice", "bob"}}
	err = repo.Save(room)
	assert.NoError(t, err)

	err = repo.RemoveUser("general", "charlie")
	assert.Error(t, err)

	err = repo.RemoveUser("general", "alice")
	assert.NoError(t, err)

	retrieved, _ := repo.Get("general")
	assert.Equal(t, 1, len(retrieved.Users))
	assert.False(t, retrieved.hasUser("alice"))
	assert.True(t, retrieved.hasUser("bob"))
}
