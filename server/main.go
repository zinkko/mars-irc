package main

import (
	"net/http"

	"github.com/gorilla/websocket"
)

func main() {
	h := createHub()
	go h.start()

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// name := r.URL.Query().Get("name")

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			panic("ai vittu!")
		}

		go acceptClient(h, conn)
	})

	http.ListenAndServe(":8080", nil)
}

func acceptClient(h *hub, conn *websocket.Conn) {
	client := &client{
		conn: conn,
	}

	conn.SetCloseHandler(func(code int, text string) error {
		h.unregister <- client
		return nil
	})

	h.register <- client

	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				continue
			}
			h.broadcast <- msg
		}
	}()
}
