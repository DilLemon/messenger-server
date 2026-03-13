package main

import (
	"log"
	"net/http"

	"messenger/internal/database"
	"messenger/internal/websocket"
)

func main() {

	database.Connect()

	http.HandleFunc("/ws", websocket.Handle)

	log.Println("server started on :8080")

	http.ListenAndServe(":8080", nil)
}
