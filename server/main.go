package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

// var planets = [2]string{"mars", "earth"}
// var distances map[string]map[string]int
// var all_hubs [2]*hub

type Configuration struct {
	Hubs      []string
	Distances map[string]map[string]int
	Meta      map[string]string
}

func main() {
	// marsHub := createHub("mars")
	// go marsHub.start()
	// earthHub := createHub("earth")
	// go earthHub.start()

	// all_hubs = [2]*hub{earthHub, marsHub}

	// distances = make(map[string]map[string]int)
	// distances["mars"] = make(map[string]int)
	// distances["earth"] = make(map[string]int)
	// distances["mars"]["mars"] = 0
	// distances["earth"]["earth"] = 0
	// distances["mars"]["earth"] = 3
	// distances["earth"]["mars"] = 3

	file, err := os.Open("server/config.json")
	if err != nil {
		panic("Failed to read config file!")
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	conf := Configuration{}
	err = decoder.Decode(&conf)
	if err != nil {
		panic("Failed to decode the config json")
	}

	fmt.Println("Starting server with the following hubs:", conf.Hubs)

	hubs := make(map[string]*hub)
	for _, name := range conf.Hubs {
		hubs[name] = createHub(name)
		go hubs[name].start()
	}

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		hub_name := r.URL.Query().Get("hub")
		if name == "" {
			name = "anon"
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Printf("%v\n\n", err)
			panic("upgrade failed!")
		}

		if hub, ok := hubs[hub_name]; ok {
			fmt.Printf("New user %s joining on %s\n", name, hub_name)
			go acceptClient(name, hub, conn, &conf, &hubs)
		} else {
			conn.Close()
			fmt.Println("User tried to join non-existent hub:", hub_name)
		}
	})

	http.ListenAndServe(":8080", nil)
}

func acceptClient(name string, h *hub, conn *websocket.Conn, conf *Configuration, hubs *map[string]*hub) {
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

		for _, receivingHub := range *hubs {
			dist := conf.Distances[h.name][receivingHub.name]
			if receivingHub == h {
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
