package models

import "time"

type Message struct {
	ID        string    `db:"id" json:"id"`
	ChatID    string    `db:"chat_id" json:"chat_id"`
	Role      string    `db:"role" json:"role"`
	Content   string    `db:"content" json:"content"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type Chat struct {
	ID        string    `db:"id" json:"id"`
	UserID    int       `db:"user_id" json:"user_id"`
	Topic     string    `db:"topic" json:"topic"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type TopicState struct {
	ID             int       `db:"id" json:"id"`
	ChatID         string    `db:"chat_id" json:"chat_id"`
	TopicText      string    `db:"topic_text" json:"topic_text"`
	TopicEmbedding []float32 `db:"topic_embedding" json:"topic_embedding"`
	Subject        string    `db:"subject" json:"subject"`
	Scope          string    `db:"scope" json:"scope"`
	Entities       []string  `db:"entities" json:"entities"`
	Locked         bool      `db:"locked" json:"locked"`
	LockedReason   string    `db:"locked_reason" json:"locked_reason"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}
