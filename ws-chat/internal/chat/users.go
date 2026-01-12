package chat

type RoomUsers []string

func (r *RoomUsers) Has(username string) bool {
	for _, user := range *r {
		if user == username {
			return true
		}
	}
	return false
}

func (r *RoomUsers) Remove(username string) {
	for i, user := range *r {
		if user == username {
			*r = append((*r)[:i], (*r)[i+1:]...)
			return
		}
	}
}
