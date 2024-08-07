package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

var (
	pongWait     = 10 * time.Second
	pingInterval = (pongWait * 9) / 10
)

type Client struct {
	connection *websocket.Conn
	manager    *Manager
	egress     chan Event
	chatRoom   string
}

type ClientList map[*Client]bool

func NewClient(m *Manager, conn *websocket.Conn) *Client {
	return &Client{connection: conn, manager: m, egress: make(chan Event), chatRoom: "general"}
}

func (c *Client) readMessage() {
	defer func() {
		c.manager.removeClient(c)
	}()

	err := c.connection.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		log.Println("error in set dead line: ", err.Error())
		return
	}
	c.connection.SetPongHandler(c.pongHandler)
	c.connection.SetReadLimit(512)
	for {
		_, payload, err := c.connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("an error occurred in socket connection : %s", err.Error())
				break
			}
		}

		var req Event
		err = json.Unmarshal(payload, &req)
		if err != nil {
			log.Println(err.Error())
			break
		}
		err = c.manager.routeEvent(req, c)
		if err != nil {
			log.Println(err.Error())
		}
	}
}

func (c *Client) writeMessage() {
	defer func() {
		c.manager.removeClient(c)
	}()

	ticker := time.NewTicker(pingInterval)
	for {
		select {
		case data, ok := <-c.egress:
			if !ok {
				err := c.connection.WriteMessage(websocket.CloseMessage, nil)
				if err != nil {
					log.Println("connection closed: ", err.Error())
				}
				return
			}
			req, err := json.Marshal(data)
			if err != nil {
				log.Println("error in marshal data ", err.Error())
				return
			}
			err = c.connection.WriteMessage(websocket.TextMessage, req)
			if err != nil {
				log.Println("error in send message: ", err.Error())
			}
			log.Println("message sent")
		case <-ticker.C:
			log.Println("ping")
			err := c.connection.WriteMessage(websocket.PingMessage, []byte(``))
			if err != nil {
				log.Println("error in ping message: ", err.Error())
				return
			}

		}

	}
}
func (c *Client) pongHandler(msg string) error {
	fmt.Println("pong")
	return c.connection.SetReadDeadline(time.Now().Add(pongWait))
}
