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
	Type string `json:"type"`
	From string `json:"from"`
	To   string `json:"to"`
	Text string `json:"text"`
	Time int64  `json:"time"`
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
		"INSERT INTO messages(from_user,to_user,text,time) VALUES($1,$2,$3,$4)",
		msg.From,
		msg.To,
		msg.Text,
		msg.Time,
	)

	if err != nil {
		log.Println(err)
	}
}

func sendHistory(user string, ws *websocket.Conn) {

	rows, err := db.Query(
		`SELECT from_user,to_user,text,time
		 FROM messages
		 WHERE from_user=$1 OR to_user=$1
		 ORDER BY id DESC LIMIT 20`,
		user,
	)

	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {

		var msg Message

		rows.Scan(&msg.From, &msg.To, &msg.Text, &msg.Time)

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

	for {

		var msg Message

		err := ws.ReadJSON(&msg)

		if err != nil {
			delete(clients, user)
			ws.Close()
			break
		}

		msg.Time = time.Now().Unix()
		msg.Type = "message"

		saveMessage(msg)

		if receiver, ok := clients[msg.To]; ok {

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
