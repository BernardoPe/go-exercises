package chat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoomMessages(t *testing.T) {
	messages := RoomMessages{}

	msg1 := Message{Sender: "alice", Content: "hello", Timestamp: 1000}
	msg2 := Message{Sender: "bob", Content: "world", Timestamp: 2000}

	messages.Add(msg1)
	messages.Add(msg2)
	assert.Equal(t, 2, len(messages))
	assert.Equal(t, "alice", messages[0].Sender)
	assert.Equal(t, "bob", messages[1].Sender)

	messages.RemoveOldest(1)
	assert.Equal(t, 1, len(messages))
	assert.Equal(t, "bob", messages[0].Sender)

	messages.RemoveOldest(10)
	assert.Equal(t, 0, len(messages))
}
