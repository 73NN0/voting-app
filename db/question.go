package db

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"gitlab.com/singfield/voting-app/entities"
)

type questionDB struct {
	ID            int
	SessionID     string
	Text          string
	OrderNum      int
	AllowMultiple int
	MaxChoices    int
	CreatedAt     Timestamp
}

func (q *questionDB) toEntity() *entities.Question {

	return &entities.Question{
		ID:            q.ID,
		SessionID:     stringToUuid(q.SessionID),
		Text:          q.Text,
		OrderNum:      q.OrderNum,
		AllowMultiple: q.AllowMultiple != 0,
		MaxChoices:    q.MaxChoices,
		CreatedAt:     q.CreatedAt.Time,
	}
}

func fromEntityQuestion(q entities.Question) *questionDB {
	var allowMultiple int
	if q.AllowMultiple {
		allowMultiple = 1
	} else {
		allowMultiple = 0
	}

	return &questionDB{
		ID:            q.ID,
		SessionID:     uuidToString(q.SessionID),
		Text:          q.Text,
		OrderNum:      q.OrderNum,
		AllowMultiple: allowMultiple,
		MaxChoices:    q.MaxChoices,
		CreatedAt:     Timestamp{q.CreatedAt},
	}
}

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

func (q *QuestionRepository) CreateQuestion(ctx context.Context, question entities.Question) (questionID int, err error) {
	qdb := fromEntityQuestion(question)

	result, err := q.db.ExecContext(ctx, `
		INSERT INTO question (session_id, text, order_num, allow_multiple, max_choices)
		VALUES (?, ?, ?, ?, ?)
	`, qdb.SessionID, qdb.Text, qdb.OrderNum, qdb.AllowMultiple, qdb.MaxChoices)

	if err != nil {
		return
	}

	id, err := result.LastInsertId()
	questionID = int(id)

	return
}

func (q *QuestionRepository) GetQuestions(ctx context.Context, sessionID uuid.UUID) ([]*entities.Question, error) {
	rows, err := q.db.QueryContext(ctx, `
		SELECT id, session_id, text, order_num, allow_multiple, max_choices, created_at
		FROM question
		WHERE session_id = ?
		ORDER BY order_num ASC
	`, uuidToString(sessionID))

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []*entities.Question
	for rows.Next() {
		q := &questionDB{}
		err := rows.Scan(&q.ID, &q.SessionID, &q.Text, &q.OrderNum, &q.AllowMultiple, &q.MaxChoices, &q.CreatedAt)
		if err != nil {
			return nil, err
		}
		questions = append(questions, q.toEntity())
	}

	return questions, rows.Err()
}

func (q *QuestionRepository) GetQuestionByID(ctx context.Context, id int) (*entities.Question, error) {
	question := questionDB{}

	err := q.db.QueryRowContext(ctx, `
		SELECT id, session_id, text, order_num, allow_multiple, max_choices, created_at
		FROM question
		WHERE id = ?
	`, id).Scan(&question.ID, &question.SessionID, &question.Text, &question.OrderNum, &question.AllowMultiple, &question.MaxChoices, &question.CreatedAt)

	if err != nil {
		return nil, err
	}

	return question.toEntity(), nil
}

func (q *QuestionRepository) DeleteQuestion(ctx context.Context, id int) error {
	_, err := q.db.ExecContext(ctx, `
		DELETE FROM question WHERE id = ?
	`, id)

	return err
}
