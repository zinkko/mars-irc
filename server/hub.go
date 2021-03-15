package main

import (
	"github.com/gorilla/websocket"
)

type client struct {
	conn *websocket.Conn
}

type hub struct {
	clients    map[*client]struct{}
	register   chan *client
	unregister chan *client
	broadcast  chan []byte
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
				client.conn.WriteMessage(websocket.TextMessage, msg)
			}
		}
	}
}

func createHub() *hub {
	return &hub{
		clients:    make(map[*client]struct{}),
		broadcast:  make(chan []byte),
		register:   make(chan *client),
		unregister: make(chan *client),
	}
}
