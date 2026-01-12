package chat

import (
	"fmt"
	"sync"
)

type InMemoryRoomRepository struct {
	rooms map[string]*Room
	sync.RWMutex
}

func NewInMemoryRoomRepository() *InMemoryRoomRepository {
	return &InMemoryRoomRepository{
		rooms: make(map[string]*Room),
	}
}

func (r *InMemoryRoomRepository) Get(name string) (*Room, bool) {
	r.RLock()
	defer r.RUnlock()
	room, exists := r.rooms[name]
	return room, exists
}

func (r *InMemoryRoomRepository) Save(room *Room) error {
	r.Lock()
	defer r.Unlock()
	r.rooms[room.Name] = room
	return nil
}

func (r *InMemoryRoomRepository) Add(roomName string, message Message) error {
	r.Lock()
	defer r.Unlock()
	room, exists := r.rooms[roomName]
	if !exists {
		room = &Room{Name: roomName}
		r.rooms[roomName] = room
	}
	room.addMessage(message)
	return nil
}

func (r *InMemoryRoomRepository) AddUser(roomName, username string) error {
	r.Lock()
	defer r.Unlock()
	room, exists := r.rooms[roomName]
	if !exists {
		room = &Room{Name: roomName}
		r.rooms[roomName] = room
	}
	if room.hasUser(username) {
		return fmt.Errorf("user %s already in room %s", username, roomName)
	}
	room.addUser(username)
	return nil
}

func (r *InMemoryRoomRepository) RemoveUser(roomName, username string) error {
	r.Lock()
	defer r.Unlock()
	room, exists := r.rooms[roomName]
	if !exists {
		return fmt.Errorf("room %s does not exist", roomName)
	}
	if !room.hasUser(username) {
		return fmt.Errorf("user %s not in room %s", username, roomName)
	}
	room.removeUser(username)
	return nil
}

func (r *InMemoryRoomRepository) GetMessages(roomName string) ([]Message, error) {
	r.RLock()
	defer r.RUnlock()
	room, exists := r.rooms[roomName]
	if !exists {
		return nil, fmt.Errorf("room %s does not exist", roomName)
	}
	messages := make([]Message, len(room.Messages))
	copy(messages, room.Messages)
	return messages, nil
}
