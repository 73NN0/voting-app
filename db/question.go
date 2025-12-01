package db

import (
	"context"
	"database/sql"
)

type QuestionRepository struct {
	db *sql.DB
}

func NewQuestionRepository(db *sql.DB) *QuestionRepository {

	if db == nil {
		panic("no db")
	}
	return &QuestionRepository{
		db: db,
	}
}

// ============= Question =============

func (q *QuestionRepository) CreateQuestion(ctx context.Context, db *sql.DB, question Question) (questionID int, err error) {
	result, err := db.ExecContext(ctx, `
		INSERT INTO question (session_id, text, order_num, allow_multiple, max_choices)
		VALUES (?, ?, ?, ?, ?)
	`, question.SessionID, question.Text, question.OrderNum, question.AllowMultiple, question.MaxChoices)

	if err != nil {
		return
	}

	id, err := result.LastInsertId()
	questionID = int(id)

	return
}

func (q *QuestionRepository) GetQuestions(ctx context.Context, db *sql.DB, sessionID string) ([]*Question, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, session_id, text, order_num, allow_multiple, max_choices, created_at
		FROM question
		WHERE session_id = ?
		ORDER BY order_num ASC
	`, sessionID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []*Question
	for rows.Next() {
		q := &Question{}
		err := rows.Scan(&q.ID, &q.SessionID, &q.Text, &q.OrderNum, &q.AllowMultiple, &q.MaxChoices, &q.CreatedAt)
		if err != nil {
			return nil, err
		}
		questions = append(questions, q)
	}

	return questions, rows.Err()
}

func (q *QuestionRepository) GetQuestionByID(ctx context.Context, db *sql.DB, id int) (*Question, error) {
	question := Question{}

	err := db.QueryRowContext(ctx, `
		SELECT id, session_id, text, order_num, allow_multiple, max_choices, created_at
		FROM question
		WHERE id = ?
	`, id).Scan(&question.ID, &question.SessionID, &question.Text, &question.OrderNum, &question.AllowMultiple, &question.MaxChoices, &question.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &question, nil
}

func (q *QuestionRepository) DeleteQuestion(ctx context.Context, db *sql.DB, id int) error {
	_, err := db.ExecContext(ctx, `
		DELETE FROM question WHERE id = ?
	`, id)

	return err
}

// ============= Choice =============
