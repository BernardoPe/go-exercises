package chat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoom(t *testing.T) {
	room := Room{Name: "general"}

	assert.False(t, room.hasUser("alice"))

	room.addUser("alice")
	assert.True(t, room.hasUser("alice"))
	assert.Equal(t, 1, len(room.Users))

	room.addUser("bob")
	assert.Equal(t, 2, len(room.Users))

	room.removeUser("alice")
	assert.False(t, room.hasUser("alice"))
	assert.Equal(t, 1, len(room.Users))

	msg := Message{Sender: "bob", Content: "test", Timestamp: 1000}
	room.addMessage(msg)
	assert.Equal(t, 1, len(room.Messages))
	assert.Equal(t, "test", room.Messages[0].Content)
}

func TestRoomMessageLimit(t *testing.T) {
	room := Room{Name: "test"}

	for i := 0; i < MessageLimit+100; i++ {
		msg := Message{Sender: "user", Content: "msg", Timestamp: int64(i)}
		room.addMessage(msg)
	}

	assert.Equal(t, MessageLimit, len(room.Messages))
	assert.Equal(t, int64(100), room.Messages[0].Timestamp)
}
