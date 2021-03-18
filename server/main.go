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
		name := r.URL.Query().Get("name")

		if name == "" {
			name = "anon"
		}

		hub := r.URL.Query().Get("hub")

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			panic("ai vittu!")
		}

		if hub == "mars" {
			fmt.Println("new user joining on Mars")
			go acceptClient(name, marsHub, conn)
		} else {
			fmt.Println("new user joining on Earth")
			go acceptClient(name, earthHub, conn)
		}
	})

	http.ListenAndServe(":8080", nil)
}

func acceptClient(name string, h *hub, conn *websocket.Conn) {
	client := &client{
		name: name,
		conn: conn,
	}

	conn.SetCloseHandler(func(code int, text string) error {
		h.unregister <- client
		return nil
	})

	h.register <- client

	go func() {
		for {
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				h.unregister <- client
				fmt.Printf("Dropping user '%s' due to error\n", name)
				break
			}
			if msgType == websocket.CloseMessage {
				fmt.Printf("%s left the server\n", name)
				h.unregister <- client
				break
			}
			if msgType != websocket.TextMessage {
				fmt.Println("unsupported message type", msgType)
				continue
			}

			for receivingHub, dist := range distances[h] {
				if dist == 0 {
					receivingHub.broadcast <- &message{name: name, data: msg}
				} else {
					go func() {
						time.Sleep(time.Duration(dist) * time.Second)
						receivingHub.broadcast <- &message{name: name, data: msg}
					}()
				}
			}
		}
	}()
}
