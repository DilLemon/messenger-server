package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
)

type Message struct {
	User string `json:"user"`
	Text string `json:"text"`
	Time int64  `json:"time"`
}

var db *sql.DB

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Message)

func connectDB() {

	connStr := "host=postgres user=messenger password=messengerpass dbname=messenger sslmode=disable"

	var err error
	db, err = sql.Open("postgres", connStr)

	if err != nil {
		log.Fatal(err)
	}

	createTable()
}

func createTable() {

	query := `
	CREATE TABLE IF NOT EXISTS messages (
		id SERIAL PRIMARY KEY,
		user_name TEXT,
		text TEXT,
		time BIGINT
	)
	`

	_, err := db.Exec(query)

	if err != nil {
		log.Fatal(err)
	}
}

func saveMessage(msg Message) {

	_, err := db.Exec(
		"INSERT INTO messages(user_name,text,time) VALUES($1,$2,$3)",
		msg.User,
		msg.Text,
		msg.Time,
	)

	if err != nil {
		log.Println(err)
	}
}

func sendHistory(ws *websocket.Conn) {

	rows, err := db.Query(
		"SELECT user_name,text,time FROM messages ORDER BY id DESC LIMIT 20",
	)

	if err != nil {
		return
	}

	defer rows.Close()

	var history []Message

	for rows.Next() {

		var m Message

		rows.Scan(&m.User, &m.Text, &m.Time)

		history = append(history, m)
	}

	for i := len(history) - 1; i >= 0; i-- {

		data, _ := json.Marshal(history[i])
		ws.WriteMessage(websocket.TextMessage, data)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {

	ws, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println(err)
		return
	}

	clients[ws] = true

	sendHistory(ws)

	for {

		var msg Message

		err := ws.ReadJSON(&msg)

		if err != nil {
			delete(clients, ws)
			ws.Close()
			break
		}

		if msg.Time == 0 {
			msg.Time = time.Now().Unix()
		}

		saveMessage(msg)

		broadcast <- msg
	}
}

func handleMessages() {

	for {

		msg := <-broadcast

		for client := range clients {

			err := client.WriteJSON(msg)

			if err != nil {

				client.Close()
				delete(clients, client)

			}
		}
	}
}

func main() {

	connectDB()

	http.HandleFunc("/ws", handleConnections)

	go handleMessages()

	log.Println("server started on :8080")

	http.ListenAndServe(":8080", nil)
}
