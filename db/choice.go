package db

import (
	"context"
	"database/sql"

	"gitlab.com/singfield/voting-app/entities"
)

type choiceDB struct {
	ID         int // AUTOINCREMENT
	QuestionID int
	Text       string
	OrderNum   int
	CreatedAt  Timestamp
}

func (c *choiceDB) toEntity() *entities.Choice {
	return &entities.Choice{
		ID:         c.ID,
		QuestionID: c.QuestionID,
		Text:       c.Text,
		OrderNum:   c.OrderNum,
		CreatedAt:  c.CreatedAt.Time,
	}
}

// I don't use pointer as argument because pointer means borrowing with possibility of mutation
func fromEntityChoice(c entities.Choice) *choiceDB {
	return &choiceDB{
		ID:         c.ID,
		QuestionID: c.QuestionID,
		Text:       c.Text,
		OrderNum:   c.OrderNum,
		CreatedAt:  Timestamp{c.CreatedAt},
	}
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

func (q *ChoiceRepository) CreateChoice(ctx context.Context, choice entities.Choice) (choiceID int, err error) {

	ch := fromEntityChoice(choice)

	result, err := q.db.ExecContext(ctx, `
		INSERT INTO choice (question_id, text, order_num)
		VALUES (?, ?, ?)
	`, ch.QuestionID, ch.Text, ch.OrderNum)

	if err != nil {
		return
	}
	id, err := result.LastInsertId()
	choiceID = int(id)

	return
}

func (q *ChoiceRepository) GetChoices(ctx context.Context, questionID int) ([]*entities.Choice, error) {
	rows, err := q.db.QueryContext(ctx, `
		SELECT id, question_id, text, order_num, created_at
		FROM choice
		WHERE question_id = ?
		ORDER BY order_num ASC
	`, questionID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var choices []*entities.Choice
	for rows.Next() {
		c := &choiceDB{}
		err := rows.Scan(&c.ID, &c.QuestionID, &c.Text, &c.OrderNum, &c.CreatedAt)
		if err != nil {
			return nil, err
		}

		choices = append(choices, c.toEntity())
	}

	return choices, rows.Err()
}

func (q *ChoiceRepository) GetChoiceByID(ctx context.Context, id int) (*entities.Choice, error) {
	c := &choiceDB{}

	err := q.db.QueryRowContext(ctx, `
		SELECT id, question_id, text, order_num, created_at
		FROM choice
		WHERE id = ?
	`, id).Scan(&c.ID, &c.QuestionID, &c.Text, &c.OrderNum, &c.CreatedAt)

	if err != nil {
		return nil, err
	}

	return c.toEntity(), nil
}

func (q *ChoiceRepository) DeleteChoice(ctx context.Context, id int) error {
	_, err := q.db.ExecContext(ctx, `
		DELETE FROM choice WHERE id = ?
	`, id)

	return err
}
