package chat

type Message struct {
	Sender    string
	Content   string
	Timestamp int64
}

type RoomMessages []Message

func (r *RoomMessages) Add(message Message) {
	*r = append(*r, message)
}

func (r *RoomMessages) RemoveOldest(n int) {
	if n >= len(*r) {
		*r = RoomMessages{}
	} else {
		*r = (*r)[n:]
	}
}
