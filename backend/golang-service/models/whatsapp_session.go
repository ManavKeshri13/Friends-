package models

import "time"

type WhatsAppQuizSession struct {
    Phone           string    `db:"phone"            json:"phone"`
    QuizID          int       `db:"quiz_id"          json:"quiz_id"`
    CurrentQuestion int       `db:"current_question" json:"current_question"`
    Score           int       `db:"score"            json:"score"`
    Active          bool      `db:"active"           json:"active"`
    CreatedAt       time.Time `db:"created_at"       json:"created_at"`
}
