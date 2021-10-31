package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var planets = [2]string{"mars", "earth"}
var distances map[string]map[string]int
var all_hubs [2]*hub

func main() {
	marsHub := createHub("mars")
	go marsHub.start()
	earthHub := createHub("earth")
	go earthHub.start()

	all_hubs = [2]*hub{earthHub, marsHub}

	distances = make(map[string]map[string]int)
	distances["mars"] = make(map[string]int)
	distances["earth"] = make(map[string]int)
	distances["mars"]["mars"] = 0
	distances["earth"]["earth"] = 0
	distances["mars"]["earth"] = 3
	distances["earth"]["mars"] = 3

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")

		if name == "" {
			name = "anon"
		}

		hub := r.URL.Query().Get("hub")

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Printf("%v\n\n", err)
			panic("upgrade failed!")
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

		for _, receivingHub := range all_hubs {
			dist := distances[h.name][receivingHub.name]
			if dist == 0 {
				receivingHub.broadcast <- &message{name: name, data: msg}
			} else {
				// take new variable, close over that
				receiver := receivingHub
				go func() {
					time.Sleep(time.Duration(dist) * time.Second)
					receiver.broadcast <- &message{name: name, data: msg}
				}()
			}
		}
	}
}
