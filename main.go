package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Message struct {
	User string `json:"user"`
	Text string `json:"text"`
	Time int64  `json:"time"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Message)

func handleConnections(w http.ResponseWriter, r *http.Request) {

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	clients[ws] = true

	for {

		var message Message

		err := ws.ReadJSON(&message)
		if err != nil {
			log.Println(err)
			delete(clients, ws)
			ws.Close()
			break
		}

		broadcast <- message
	}
}

func handleMessages() {

	for {

		msg := <-broadcast

		for client := range clients {

			err := client.WriteJSON(msg)
			if err != nil {
				log.Println(err)
				client.Close()
				delete(clients, client)
			}

		}

	}
}

func main() {

	http.HandleFunc("/ws", handleConnections)

	go handleMessages()

	log.Println("server started on :8080")

	http.ListenAndServe(":8080", nil)

}
