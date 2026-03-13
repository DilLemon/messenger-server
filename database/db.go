package database

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"messenger/models"
)

var DB *sql.DB

func Connect() {

	connStr := "host=postgres user=messenger password=messengerpass dbname=messenger sslmode=disable"

	var err error
	DB, err = sql.Open("postgres", connStr)

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
		status TEXT
	)
	`

	DB.Exec(query)
}

func SaveMessage(msg models.Message) {

	_, err := DB.Exec(
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

func LoadHistory(user string) ([]models.Message, error) {

	rows, err := DB.Query(
		`SELECT from_user,to_user,text,time,status
		 FROM messages
		 WHERE from_user=$1 OR to_user=$1
		 ORDER BY id DESC LIMIT 20`,
		user,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []models.Message

	for rows.Next() {

		var m models.Message

		rows.Scan(&m.From, &m.To, &m.Text, &m.Time, &m.Status)

		m.Type = "message"

		result = append(result, m)
	}

	return result, nil
}
