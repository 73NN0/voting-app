package db

import (
	"context"
	"database/sql"
)

type Choice struct {
	ID         int // AUTOINCREMENT
	QuestionID int
	Text       string
	OrderNum   int
	CreatedAt  Timestamp
}

type ChoiceRepository struct {
	db *sql.DB
}

func NewChoiceRepository(db *sql.DB) *ChoiceRepository {
	if db == nil {
		panic("no db")
	}

	return &ChoiceRepository{
		db: db,
	}
}

func (q *ChoiceRepository) CreateChoice(ctx context.Context, db *sql.DB, choice Choice) (choiceID int, err error) {
	result, err := db.ExecContext(ctx, `
		INSERT INTO choice (question_id, text, order_num)
		VALUES (?, ?, ?)
	`, choice.QuestionID, choice.Text, choice.OrderNum)

	if err != nil {
		return
	}
	id, err := result.LastInsertId()
	choiceID = int(id)

	return
}

func (q *ChoiceRepository) GetChoices(ctx context.Context, db *sql.DB, questionID int) ([]*Choice, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, question_id, text, order_num, created_at
		FROM choice
		WHERE question_id = ?
		ORDER BY order_num ASC
	`, questionID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var choices []*Choice
	for rows.Next() {
		c := &Choice{}
		err := rows.Scan(&c.ID, &c.QuestionID, &c.Text, &c.OrderNum, &c.CreatedAt)
		if err != nil {
			return nil, err
		}
		choices = append(choices, c)
	}

	return choices, rows.Err()
}

func (q *ChoiceRepository) GetChoiceByID(ctx context.Context, db *sql.DB, id int) (*Choice, error) {
	c := &Choice{}

	err := db.QueryRowContext(ctx, `
		SELECT id, question_id, text, order_num, created_at
		FROM choice
		WHERE id = ?
	`, id).Scan(&c.ID, &c.QuestionID, &c.Text, &c.OrderNum, &c.CreatedAt)

	if err != nil {
		return nil, err
	}

	return c, nil
}

func (q *ChoiceRepository) DeleteChoice(ctx context.Context, db *sql.DB, id int) error {
	_, err := db.ExecContext(ctx, `
		DELETE FROM choice WHERE id = ?
	`, id)

	return err
}
