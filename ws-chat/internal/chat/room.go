package chat

const (
	MessageLimit = 1000
)

type Room struct {
	Name     string
	Users    RoomUsers
	Messages RoomMessages
}

func (r *Room) hasUser(username string) bool {
	return r.Users.Has(username)
}

func (r *Room) removeUser(username string) {
	r.Users.Remove(username)
}

func (r *Room) addUser(username string) {
	r.Users = append(r.Users, username)
}

func (r *Room) addMessage(message Message) {
	r.Messages.Add(message)
	if len(r.Messages) > MessageLimit {
		r.removeOldMessages(len(r.Messages) - MessageLimit)
	}
}

func (r *Room) removeOldMessages(n int) {
	r.Messages.RemoveOldest(n)
}
