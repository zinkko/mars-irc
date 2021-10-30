package main

import (
	"fmt"

	"github.com/gorilla/websocket"
)

type client struct {
	name   string
	origin string
	conn   *websocket.Conn
}

type message struct {
	name   string
	origin string
	data   []byte
}

type hub struct {
	name       string
	clients    map[*client]struct{}
	register   chan *client
	unregister chan *client
	broadcast  chan *message
}

func (h *hub) start() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = struct{}{}
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
			}
		case msg := <-h.broadcast:
			fmt.Printf("broadcast %s on hub(%s)\n", string(msg.data), h.name)
			for client := range h.clients {
				fmt.Printf("\tclient: %v\n", client)
				if msg.name != client.name {
					fmt.Printf("\t\tSending message thru socket: %s\n", string(msg.data))
					prefix := fmt.Sprintf("%s: ", msg.name)
					client.conn.WriteMessage(websocket.TextMessage, append([]byte(prefix), msg.data...))
				}
			}
		}
	}
}

func createHub(name string) *hub {
	return &hub{
		name:       name,
		clients:    make(map[*client]struct{}),
		broadcast:  make(chan *message),
		register:   make(chan *client),
		unregister: make(chan *client),
	}
}
