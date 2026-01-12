package chat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoomUsers(t *testing.T) {
	users := RoomUsers{"alice", "bob", "charlie"}

	assert.True(t, users.Has("alice"))
	assert.True(t, users.Has("bob"))
	assert.False(t, users.Has("dave"))

	users.Remove("bob")
	assert.False(t, users.Has("bob"))
	assert.Equal(t, 2, len(users))

	users.Remove("nonexistent")
	assert.Equal(t, 2, len(users))
}
