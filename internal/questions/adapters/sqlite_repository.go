package adapters

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/73NN0/voting-app/internal/common/db"
	choice "github.com/73NN0/voting-app/internal/questions/domain/choice"
	"github.com/google/uuid"

	"github.com/73NN0/voting-app/internal/questions/domain/question"
)

// ======================= DTO ==================== //
type choiceDTO struct {
	ID         int          // INTEGER AUTOINCREMENT
	QuestionID int          // INTEGER
	Text       string       // TEXT
	OrderNum   int          // INTEGER
	CreatedAt  db.Timestamp // TEXT
}

func toChoiceDTO(c *choice.Choice) choiceDTO {
	return choiceDTO{
		ID:         c.ID(),
		QuestionID: c.QuestionID(),
		Text:       c.Text(),
		OrderNum:   c.OrderNum(),
		CreatedAt:  db.Timestamp{Time: c.CreatedAt()},
	}
}

func (dto choiceDTO) toDomain() (*choice.Choice, error) {
	return choice.UnmarshalChoiceFromRepository(
		dto.ID,
		dto.QuestionID,
		dto.Text,
		dto.OrderNum,
		dto.CreatedAt.Time,
	)
}

type questionDTO struct {
	ID            int          // INTEGER AUTOINCREMENT
	SessionID     string       // TEXT (uuid en string)
	Text          string       // TEXT
	OrderNum      int          // INTEGER
	AllowMultiple int          // INTEGER (0 ou 1, SQLite n'a pas de bool natif)
	MaxChoices    int          // INTEGER
	CreatedAt     db.Timestamp // TEXT (format ISO)
}

func toQuestionDTO(s *question.Question) questionDTO {
	var allowMultiple int
	if s.AllowMultiple() {
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

func (dto questionDTO) toDomain() (*question.Question, error) {
	sessionID, err := uuid.Parse(dto.SessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session_id uuid: %w", err)
	}

	allowMultiple := dto.AllowMultiple != 0

	return question.UnmarshalQuestionFromRepository(
		dto.ID,
		sessionID,
		dto.Text,
		dto.OrderNum,
		allowMultiple,
		dto.MaxChoices,
		dto.CreatedAt.Time,
	)
}

type SqliteQuestionsRepository struct {
	dbRepo db.DBRepository
}

func NewSqliteQuestionsRepository(dbrepo db.DBRepository) *SqliteQuestionsRepository {

	return &SqliteQuestionsRepository{dbRepo: dbrepo}
}

// ================= Questions ======================= //
func (s *SqliteQuestionsRepository) CreateQuestion(ctx context.Context, question *question.Question) (questionID int, err error) {
	if question == nil {
		err = fmt.Errorf("question ptr is nil !")
		return
	}

	dto := toQuestionDTO(question)

	result, err := s.dbRepo.ExecContext(ctx, `
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

func (r *SqliteQuestionsRepository) GetQuestionByID(ctx context.Context, id int) (*question.Question, error) {
	var dto questionDTO

	err := r.dbRepo.QueryRowContext(ctx, `
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
			return nil, fmt.Errorf("question %d not found", id)
		}
		return nil, fmt.Errorf("failed to query question: %w", err)
	}

	return dto.toDomain()
}

func (r *SqliteQuestionsRepository) GetQuestionsBySessionID(ctx context.Context, sessionID uuid.UUID) ([]*question.Question, error) {
	rows, err := r.dbRepo.QueryContext(ctx, `
		SELECT id, session_id, text, order_num, allow_multiple, max_choices, created_at
		FROM question
		WHERE session_id = ?
		ORDER BY order_num ASC
	`, sessionID.String())

	if err != nil {
		return nil, fmt.Errorf("failed to query questions: %w", err)
	}
	defer rows.Close()

	var questions []*question.Question
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

		q, err := dto.toDomain()
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
	_, err := r.dbRepo.ExecContext(ctx, `DELETE FROM question WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete question %d: %w", id, err)
	}
	return nil
}

// =================== Choices ======================= //

func (r *SqliteQuestionsRepository) CreateChoice(ctx context.Context, c *choice.Choice) (int, error) {
	dto := toChoiceDTO(c)

	result, err := r.dbRepo.ExecContext(ctx, `
		INSERT INTO choice (question_id, text, order_num)
		VALUES (?, ?, ?)
	`, dto.QuestionID, dto.Text, dto.OrderNum)

	if err != nil {
		return 0, fmt.Errorf("failed to insert choice: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return int(id), nil
}

func (r *SqliteQuestionsRepository) GetChoiceByID(ctx context.Context, id int) (*choice.Choice, error) {
	var dto choiceDTO

	err := r.dbRepo.QueryRowContext(ctx, `
		SELECT id, question_id, text, order_num, created_at
		FROM choice
		WHERE id = ?
	`, id).Scan(
		&dto.ID,
		&dto.QuestionID,
		&dto.Text,
		&dto.OrderNum,
		&dto.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("choice %d not found", id)
		}
		return nil, fmt.Errorf("failed to query choice: %w", err)
	}

	return dto.toDomain()
}

func (r *SqliteQuestionsRepository) GetChoicesByQuestionID(ctx context.Context, questionID int) ([]*choice.Choice, error) {
	rows, err := r.dbRepo.QueryContext(ctx, `
		SELECT id, question_id, text, order_num, created_at
		FROM choice
		WHERE question_id = ?
		ORDER BY order_num ASC
	`, questionID)

	if err != nil {
		return nil, fmt.Errorf("failed to query choices: %w", err)
	}
	defer rows.Close()

	var choices []*choice.Choice
	for rows.Next() {
		var dto choiceDTO
		err := rows.Scan(
			&dto.ID,
			&dto.QuestionID,
			&dto.Text,
			&dto.OrderNum,
			&dto.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan choice row: %w", err)
		}

		c, err := dto.toDomain()
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal choice: %w", err)
		}

		choices = append(choices, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating choice rows: %w", err)
	}

	return choices, nil
}

func (q *SqliteQuestionsRepository) DeleteChoice(ctx context.Context, choiceID int) error {
	_, err := q.dbRepo.ExecContext(ctx, `
		DELETE FROM choice WHERE id = ?
	`, choiceID)

	return err
}
