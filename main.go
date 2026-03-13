package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
)

type Message struct {
	Type   string `json:"type"`
	From   string `json:"from"`
	To     string `json:"to"`
	Text   string `json:"text"`
	Time   int64  `json:"time"`
	Status string `json:"status"`
}

type Login struct {
	Type string `json:"type"`
	User string `json:"user"`
}

var db *sql.DB

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var clients = map[string]*websocket.Conn{}

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
		from_user TEXT,
		to_user TEXT,
		text TEXT,
		time BIGINT,
		status TEXT DEFAULT 'sent'
	)
	`

	_, err := db.Exec(query)

	if err != nil {
		log.Fatal(err)
	}
}

func saveMessage(msg Message) {

	_, err := db.Exec(
		"INSERT INTO messages(from_user,to_user,text,time,status) VALUES($1,$2,$3,$4,$5)",
		msg.From,
		msg.To,
		msg.Text,
		msg.Time,
		msg.Status,
	)

	if err != nil {
		log.Println(err)
	}
}

func markDelivered(user string) {

	_, err := db.Exec(
		"UPDATE messages SET status='delivered' WHERE to_user=$1 AND status='sent'",
		user,
	)

	if err != nil {
		log.Println(err)
	}
}

func sendOfflineMessages(user string, ws *websocket.Conn) {

	rows, err := db.Query(
		`SELECT from_user,to_user,text,time,status
		 FROM messages
		 WHERE to_user=$1 AND status='sent'
		 ORDER BY id ASC`,
		user,
	)

	if err != nil {
		log.Println(err)
		return
	}

	defer rows.Close()

	for rows.Next() {

		var msg Message

		rows.Scan(&msg.From, &msg.To, &msg.Text, &msg.Time, &msg.Status)

		msg.Type = "message"

		ws.WriteJSON(msg)
	}

	markDelivered(user)
}

func sendHistory(user string, ws *websocket.Conn) {

	rows, err := db.Query(
		`SELECT from_user,to_user,text,time,status
		 FROM messages
		 WHERE from_user=$1 OR to_user=$1
		 ORDER BY id DESC LIMIT 20`,
		user,
	)

	if err != nil {
		log.Println(err)
		return
	}

	defer rows.Close()

	for rows.Next() {

		var msg Message

		rows.Scan(&msg.From, &msg.To, &msg.Text, &msg.Time, &msg.Status)

		msg.Type = "message"

		ws.WriteJSON(msg)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {

	ws, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println(err)
		return
	}

	var login Login

	err = ws.ReadJSON(&login)

	if err != nil || login.Type != "login" {
		ws.Close()
		return
	}

	user := login.User

	clients[user] = ws

	log.Println("user connected:", user)

	sendHistory(user, ws)
	sendOfflineMessages(user, ws)

	for {

		var msg Message

		err := ws.ReadJSON(&msg)

		if err != nil {
			delete(clients, user)
			ws.Close()
			log.Println("user disconnected:", user)
			break
		}

		msg.Time = time.Now().Unix()
		msg.Type = "message"
		msg.Status = "sent"

		saveMessage(msg)

		if receiver, ok := clients[msg.To]; ok {

			msg.Status = "delivered"

			receiver.WriteJSON(msg)
		}

		if sender, ok := clients[msg.From]; ok {

			sender.WriteJSON(msg)
		}
	}
}

func main() {

	connectDB()

	http.HandleFunc("/ws", handleConnections)

	log.Println("server started on :8080")

	http.ListenAndServe(":8080", nil)
}
