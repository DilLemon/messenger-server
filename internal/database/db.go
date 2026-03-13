package database

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
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
