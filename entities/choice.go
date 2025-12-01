package entities

import "time"

type Choice struct {
	ID         int // AUTOINCREMENT
	QuestionID int
	Text       string
	OrderNum   int
	CreatedAt  time.Time
}
