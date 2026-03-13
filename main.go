package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan string)

func handleConnections(w http.ResponseWriter, r *http.Request) {

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	clients[ws] = true

	for {
		_, msg, err := ws.ReadMessage()

		if err != nil {
			delete(clients, ws)
			break
		}

		broadcast <- string(msg)
	}
}

func handleMessages() {
	for {
		msg := <-broadcast

		for client := range clients {
			client.WriteMessage(websocket.TextMessage, []byte(msg))
		}
	}
}

func main() {

	http.HandleFunc("/ws", handleConnections)

	go handleMessages()

	log.Println("server started on :8080")

	http.ListenAndServe(":8080", nil)
}
