package main

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

var ClientList map[*Client]bool

type Client struct {
	connection *websocket.Conn
	manager    *Manager
	egress     chan []byte
}

func NewClient(m *Manager, conn *websocket.Conn) *Client {
	return &Client{connection: conn, manager: m, egress: make(chan []byte)}
}

func (c *Client) readMessage() {
	defer func() {
		c.manager.removeClient(c)
	}()
	for {
		messageType, message, err := c.connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("an error occurred in socket connection : %s", err.Error())
				break
			}
		}
		fmt.Println(messageType)
		fmt.Println(string(message))

		for v, _ := range c.manager.clients {
			v.egress <- message
		}
	}
}

func (c *Client) writeMessage() {
	defer func() {
		c.manager.removeClient(c)
	}()

	for {
		select {
		case message, ok := <-c.egress:
			if !ok {
				err := c.connection.WriteMessage(websocket.CloseMessage, nil)
				if err != nil {
					log.Println("connection closed: ", err.Error())
				}
				return
			}
			err := c.connection.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Println("error in send message: ", err.Error())
			}
			log.Println("message sent")

		}

	}
}
