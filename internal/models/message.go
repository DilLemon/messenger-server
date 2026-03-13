package models

type Message struct {
	ID     string `json:"id"`
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

type Receipt struct {
	Type      string `json:"type"`
	MessageID string `json:"messageId"`
	User      string `json:"user"`
}
