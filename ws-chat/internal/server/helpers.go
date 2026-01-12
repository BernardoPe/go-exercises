package server

import "sync"

type roomSet map[string]struct{}

func (rs roomSet) Add(roomName string)    { rs[roomName] = struct{}{} }
func (rs roomSet) Remove(roomName string) { delete(rs, roomName) }
func (rs roomSet) Has(roomName string) bool {
	_, exists := rs[roomName]
	return exists
}

type Clients map[*Client]struct{}

func (c Clients) Add(client *Client)    { c[client] = struct{}{} }
func (c Clients) Remove(client *Client) { delete(c, client) }
func (c Clients) ForEach(fn func(*Client)) {
	for client := range c {
		fn(client)
	}
}

type usernameGenerator struct {
	mu sync.Mutex
	n  int
}

// Next generates the next unique username in the format "user_X" where X is an incrementing integer.
// In a real application we would have a proper authentication system.
// Here the main focus is on WebSocket communication, so I use this simple generator.
func (g *usernameGenerator) Next() string {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.n++
	return "user_" + itoa(g.n)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
