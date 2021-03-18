package main

import (
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
			for client := range h.clients {
				if msg.name != client.name {
					client.conn.WriteMessage(websocket.TextMessage, msg.data)
				}
			}
		}
	}
}

func createHub() *hub {
	return &hub{
		clients:    make(map[*client]struct{}),
		broadcast:  make(chan *message),
		register:   make(chan *client),
		unregister: make(chan *client),
	}
}
