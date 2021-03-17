package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var planets = [2]string{"mars", "earth"}
var distances map[*hub]map[*hub]int

func main() {
	marsHub := createHub()
	go marsHub.start()
	earthHub := createHub()
	go earthHub.start()

	distances = make(map[*hub]map[*hub]int)
	distances[marsHub] = make(map[*hub]int)
	distances[earthHub] = make(map[*hub]int)
	distances[marsHub][marsHub] = 0
	distances[earthHub][earthHub] = 0
	distances[marsHub][earthHub] = 3
	distances[earthHub][marsHub] = 3

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// name := r.URL.Query().Get("name")
		hub := r.URL.Query().Get("hub")

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			panic("ai vittu!")
		}

		if hub == "mars" {
			fmt.Println("new user joining on Mars")
			go acceptClient(marsHub, conn)
		} else {
			fmt.Println("new user joining on Earth")
			go acceptClient(earthHub, conn)
		}
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
			for receivingHub, dist := range distances[h] {
				if dist == 0 {
					receivingHub.broadcast <- msg
				} else {
					go func() {
						time.Sleep(time.Duration(dist) * time.Second)
						receivingHub.broadcast <- msg
					}()
				}
			}
		}
	}()
}
