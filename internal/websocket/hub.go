package websocket

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"messenger/internal/database"
	"messenger/internal/models"
)

var clients = map[string]*websocket.Conn{}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func Handle(w http.ResponseWriter, r *http.Request) {

	ws, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println(err)
		return
	}

	var login models.Login

	err = ws.ReadJSON(&login)

	if err != nil {
		ws.Close()
		return
	}

	user := login.User

	clients[user] = ws

	log.Println("connected:", user)

	for {

		var msg models.Message

		err := ws.ReadJSON(&msg)

		if err != nil {

			delete(clients, user)
			ws.Close()

			log.Println("disconnected:", user)

			break
		}

		msg.Time = time.Now().Unix()
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

func saveMessage(msg models.Message) {

	_, err := database.DB.Exec(
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
