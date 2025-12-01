package entities

import "gitlab.com/singfield/voting-app/db"

type Choice struct {
	ID         int // AUTOINCREMENT
	QuestionID int
	Text       string
	OrderNum   int
	CreatedAt  db.Timestamp // NOT OK WITH THAT
}
