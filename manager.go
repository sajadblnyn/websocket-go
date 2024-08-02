package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	websocketUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type Manager struct {
	clients map[*Client]bool
	sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{clients: map[*Client]bool{}}
}

func (m *Manager) ServeWs(w http.ResponseWriter, r *http.Request) {
	log.Println("new connection")

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
