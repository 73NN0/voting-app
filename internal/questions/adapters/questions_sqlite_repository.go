package adapters

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/73NN0/voting-app/internal/common/db"
	"github.com/google/uuid"

	"github.com/73NN0/voting-app/internal/questions/domain/question"
)

// ======================= DTO ==================== //

type questionDTO struct {
	ID            int          // INTEGER AUTOINCREMENT
	SessionID     string       // TEXT (uuid en string)
	Text          string       // TEXT
	OrderNum      int          // INTEGER
	AllowMultiple int          // INTEGER (0 or 1, SQLite doesn't have a natif boolean)
	MaxChoices    int          // INTEGER
	CreatedAt     db.Timestamp // TEXT (format ISO)
}

func toQuestionDTO(s *question.Question) questionDTO {
	var allowMultiple int
	if s.AllowMultiple() { // convertion
		allowMultiple = 1
	}

	return questionDTO{
		ID:            s.ID(), // 0 pour un INSERT
		SessionID:     s.SessionID().String(),
		Text:          s.Text(),
		OrderNum:      s.OrderNum(),
		AllowMultiple: allowMultiple,
		MaxChoices:    s.MaxChoices(),
		CreatedAt:     db.Timestamp{Time: s.CreatedAt()},
	}
}

func (dto questionDTO) toQuestion() (question.Question, error) {
	sessionID, err := uuid.Parse(dto.SessionID)
	if err != nil {
		return question.Question{}, fmt.Errorf("invalid session_id uuid: %w", err)
	}

	allowMultiple := dto.AllowMultiple != 0

	ptr, err := question.Rehydrate(
		dto.ID,
		sessionID,
		dto.Text,
		dto.OrderNum,
		allowMultiple,
		dto.MaxChoices,
		dto.CreatedAt.Time,
	)

	return *ptr, err
}

type SqliteQuestionsRepository struct {
	db *sql.DB
}

func NewSqliteQuestionsRepository(db *sql.DB) *SqliteQuestionsRepository {

	return &SqliteQuestionsRepository{db: db}
}

// ================= Questions ======================= //
func (s *SqliteQuestionsRepository) CreateQuestion(ctx context.Context, question question.Question) (questionID int, err error) {
	// TODO indepotent ?
	dto := toQuestionDTO(&question)

	result, err := s.db.ExecContext(ctx, `
		INSERT INTO question (session_id, text, order_num, allow_multiple, max_choices)
		VALUES (?, ?, ?, ?, ?)
	`, dto.SessionID, dto.Text, dto.OrderNum, dto.AllowMultiple, dto.MaxChoices)

	if err != nil {
		return
	}

	id, err := result.LastInsertId()
	questionID = int(id)

	return
}

func (r *SqliteQuestionsRepository) GetQuestionByID(ctx context.Context, id int) (question.Question, error) {
	var dto questionDTO

	err := r.db.QueryRowContext(ctx, `
		SELECT id, session_id, text, order_num, allow_multiple, max_choices, created_at
		FROM question
		WHERE id = ?
	`, id).Scan(
		&dto.ID,
		&dto.SessionID,
		&dto.Text,
		&dto.OrderNum,
		&dto.AllowMultiple,
		&dto.MaxChoices,
		&dto.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return question.Question{}, fmt.Errorf("question %d not found", id)
		}
		return question.Question{}, fmt.Errorf("failed to query question: %w", err)
	}

	return dto.toQuestion()
}

func (r *SqliteQuestionsRepository) GetQuestionsBySessionID(ctx context.Context, sessionID uuid.UUID) ([]question.Question, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, session_id, text, order_num, allow_multiple, max_choices, created_at
		FROM question
		WHERE session_id = ?
		ORDER BY order_num ASC
	`, sessionID.String())

	if err != nil {
		return nil, fmt.Errorf("failed to query questions: %w", err)
	}
	defer rows.Close()

	var questions []question.Question
	for rows.Next() {
		var dto questionDTO
		err := rows.Scan(
			&dto.ID,
			&dto.SessionID,
			&dto.Text,
			&dto.OrderNum,
			&dto.AllowMultiple,
			&dto.MaxChoices,
			&dto.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan question row: %w", err)
		}

		q, err := dto.toQuestion()
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal question: %w", err)
		}

		questions = append(questions, q)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating question rows: %w", err)
	}

	return questions, nil
}

func (r *SqliteQuestionsRepository) DeleteQuestion(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM question WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete question %d: %w", id, err)
	}
	return nil
}

func (r *SqliteQuestionsRepository) UpdateQuestion(ctx context.Context, q question.Question) error {

	if _, err := r.db.ExecContext(ctx, `
		UPDATE question
		SET text = ?, order_num = ?, allow_multiple = ?, max_choices = ?
		WHERE id = ?
	`, q.Text(), q.OrderNum(), q.AllowMultiple(), q.MaxChoices(), q.ID()); err != nil {
		return fmt.Errorf("failed to update question %d : %w", q.ID(), err)
	}

	return nil
}
