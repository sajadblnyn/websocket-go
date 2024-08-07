package main

import (
	"encoding/json"
	"time"
)

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type EventHandler func(event Event, client *Client) error

const (
	SendMessageEventType    string = "send_message"
	NewMessageEventType     string = "new_message"
	ChangeChatRoomEventType string = "change_room"
)

type SendMessagePayload struct {
	Message string `json:"message"`
	From    string `json:"from"`
}

type NewMessageEvent struct {
	SendMessagePayload
	SentAt time.Time `json:"sentAt"`
}

type ChangeRoomPayload struct {
	Name string `json:"name"`
}
