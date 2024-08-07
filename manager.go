package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	websocketUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     checkOrigin,
	}
)

type Manager struct {
	clients ClientList
	sync.RWMutex
	handlers map[string]EventHandler

	otps RetentionMap
}

func NewManager(ctx context.Context) *Manager {
	m := &Manager{clients: ClientList{}, handlers: make(map[string]EventHandler), otps: NewRetentionMap(ctx, time.Second*5)}
	m.InitialHandlers()
	return m
}

func (m *Manager) LoginHandler(w http.ResponseWriter, r *http.Request) {

	type LoginForm struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var form LoginForm

	err := json.NewDecoder(r.Body).Decode(&form)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if form.Username != "sajad" || form.Password != "sajad" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	type Response struct {
		Otp string `json:"otp"`
	}
	o := m.otps.NewOTP()
	var res Response = Response{Otp: o.Key}

	jsonRes, err := json.Marshal(res)
	if form.Username != "sajad" || form.Password != "sajad" {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonRes)
}

func (m *Manager) InitialHandlers() {
	m.handlers = map[string]EventHandler{
		SendMessageEventType:    SendMessage,
		ChangeChatRoomEventType: ChangeChatRoomHandler,
	}
}

func ChangeChatRoomHandler(event Event, client *Client) error {
	var changeRoomEvent ChangeRoomPayload

	err := json.Unmarshal(event.Payload, &changeRoomEvent)
	if err != nil {
		return err
	}
	client.chatRoom = changeRoomEvent.Name
	return nil
}

func SendMessage(event Event, client *Client) error {
	var sendMessagePayload SendMessagePayload

	err := json.Unmarshal(event.Payload, &sendMessagePayload)
	if err != nil {
		return err
	}

	var newMessageEvent NewMessageEvent
	newMessageEvent.SentAt = time.Now()
	newMessageEvent.From = sendMessagePayload.From
	newMessageEvent.Message = sendMessagePayload.Message

	data, err := json.Marshal(newMessageEvent)

	if err != nil {
		return err
	}

	outputEvent := Event{
		Type:    NewMessageEventType,
		Payload: data,
	}

	for v, _ := range client.manager.clients {
		if v.chatRoom == client.chatRoom {
			v.egress <- outputEvent
		}
	}
	return nil

}

func (m *Manager) routeEvent(event Event, c *Client) error {
	handler, ok := m.handlers[event.Type]
	if !ok {
		return errors.New("no valid handler found by this event type")
	}
	err := handler(event, c)
	if err != nil {
		return err
	}
	return nil
}

func (m *Manager) ServeWs(w http.ResponseWriter, r *http.Request) {
	log.Println("new connection")

	otp := r.URL.Query().Get("otp")

	if otp == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if !m.otps.Verify(otp) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	//upgrade regular http connection to websocket
	conn, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err.Error())
	}
	client := NewClient(m, conn)
	m.addClient(client)

	go client.readMessage()
	go client.writeMessage()

}

func (m *Manager) addClient(conn *Client) {
	m.Lock()
	defer m.Unlock()
	m.clients[conn] = true
}

func (m *Manager) removeClient(conn *Client) {
	m.Lock()
	defer m.Unlock()
	_, exists := m.clients[conn]
	if exists {
		conn.connection.Close()
		delete(m.clients, conn)
	}
}

func checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	switch origin {
	case "http://127.0.0.1:8080":
		return true
	case "http://localhost:8080":
		return true
	default:
		return false
	}
}
