package websocket

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"messenger/database"
	"messenger/models"
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

	history, _ := database.LoadHistory(user)

	for i := len(history) - 1; i >= 0; i-- {

		ws.WriteJSON(history[i])
	}

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

		database.SaveMessage(msg)

		if receiver, ok := clients[msg.To]; ok {

			msg.Status = "delivered"

			receiver.WriteJSON(msg)
		}

		if sender, ok := clients[msg.From]; ok {

			sender.WriteJSON(msg)
		}
	}
}
