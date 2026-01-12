package chat

type RoomRepository interface {
	Get(name string) (*Room, bool)
	Save(room *Room) error
	Add(roomName string, message Message) error
	AddUser(roomName, username string) error
	RemoveUser(roomName, username string) error
	GetMessages(roomName string) ([]Message, error)
}
