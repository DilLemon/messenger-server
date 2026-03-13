package main

import (
	"log"
	"net/http"

	"messenger/database"
	"messenger/websocket"
)

func main() {

	database.Connect()

	http.HandleFunc("/ws", websocket.Handle)

	log.Println("server started on :8080")

	http.ListenAndServe(":8080", nil)
}
