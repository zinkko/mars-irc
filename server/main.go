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
			// fmt.Println("Checking requester origin...")
			// fmt.Printf("%s", r)

			// parts := strings.Split(r.URL.Host, ":")
			// if len(parts) == 0 {
			// 	return false
			// }
			// return parts[0] == "127.0.0.1"
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

			for _, receivingHub := range all_hubs {
				dist := distances[h.name][receivingHub.name]

				fmt.Printf("client of %s, send to: %v\n", h.name, receivingHub.name)
				if dist == 0 {
					fmt.Printf("(%s) broadcast within hub(%s): %s sends: %s\n", h.name, receivingHub.name, name, string(msg))
					receivingHub.broadcast <- &message{name: name, data: msg}
				} else {
					go func() {
						time.Sleep(time.Duration(dist) * time.Second)
						fmt.Printf("(%s) broadcast to OTHER hub(%s): %s sends: %s\n", h.name, receivingHub.name, name, string(msg))
						receivingHub.broadcast <- &message{name: name, data: msg}
					}()
				}
			}
		}
	}()
}
