package models

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
