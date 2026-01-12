package model

import "encoding/json"

type Command int

const (
	CmdSendMessage Command = iota + 1
	CmdJoinRoom
	CmdLeaveRoom
)

type WsCommand struct {
	Command Command
	Payload json.RawMessage
}

func NewWsCommand(cmd Command, payload json.RawMessage) *WsCommand {
	return &WsCommand{
		Command: cmd,
		Payload: payload,
	}
}

type JoinRoomPayload struct {
	RoomName string `json:"roomName"`
}

type LeaveRoomPayload struct {
	RoomName string `json:"roomName"`
}

type SendMessagePayload struct {
	RoomName string `json:"roomName"`
	Content  string `json:"content"`
}
