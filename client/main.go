package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

func prompt(text string) string {
	var s string
	fmt.Print(text)
	fmt.Scanln(&s)
	return strings.TrimSpace(s)
}

func main() {
	name := prompt("What is your name? ")
	hubName := prompt("Which hub would you like to join? ")

	hubName = strings.TrimSpace(hubName)
	url := fmt.Sprintf("ws://localhost:8080?hub=%s&name=%s", hubName, name)
	conn, _, _ := websocket.DefaultDialer.Dial(url, nil)

	defer conn.Close()

	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				os.Exit(1)
			}
			fmt.Printf("%s\n", msg)
		}
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	go func() {
		<-interrupt
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		fmt.Println("forcefully exiting...")
		time.Sleep(3 * time.Second)
		os.Exit(1)
	}()

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("> ")
		msg, _ := reader.ReadString('\n')
		msg = strings.TrimSpace(msg)
		if msg != "" {
			conn.WriteMessage(websocket.TextMessage, []byte(msg))
		}
	}
}
