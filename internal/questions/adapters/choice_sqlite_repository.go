package adapters

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/73NN0/voting-app/internal/common/db"
	"github.com/73NN0/voting-app/internal/questions/domain/choice"
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

func (dto choiceDTO) toChoice() (choice.Choice, error) {
	// in this context we don't need to send a pointer
	ptr, err := choice.Rehydrate(
		dto.ID,
		dto.QuestionID,
		dto.Text,
		dto.OrderNum,
		dto.CreatedAt.Time,
	)

	return *ptr, err
}

//

type SqliteChoicesRepository struct {
	db *sql.DB
}

func NewSqliteChoicesRepositoy(db *sql.DB) *SqliteChoicesRepository {

	if db == nil {
		panic("no db in SQL choice repository !")
	}

	return &SqliteChoicesRepository{
		db: db,
	}
}

func (r *SqliteChoicesRepository) CreateChoice(ctx context.Context, c choice.Choice) (int, error) {
	dto := toChoiceDTO(&c)

	result, err := r.db.ExecContext(ctx, `
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

func (r *SqliteChoicesRepository) GetChoiceByID(ctx context.Context, id int) (choice.Choice, error) {
	var dto choiceDTO

	err := r.db.QueryRowContext(ctx, `
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
			return choice.Choice{}, fmt.Errorf("choice %d not found", id)
		}
		return choice.Choice{}, fmt.Errorf("failed to query choice: %w", err)
	}

	return dto.toChoice()
}

func (r *SqliteChoicesRepository) GetChoicesByQuestionID(ctx context.Context, questionID int) ([]choice.Choice, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, question_id, text, order_num, created_at
		FROM choice
		WHERE question_id = ?
		ORDER BY order_num ASC
	`, questionID)

	if err != nil {
		return nil, fmt.Errorf("failed to query choices: %w", err)
	}
	defer rows.Close()

	var choices []choice.Choice
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

		c, err := dto.toChoice()
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

func (q *SqliteChoicesRepository) DeleteChoice(ctx context.Context, choiceID int) error {
	_, err := q.db.ExecContext(ctx, `
		DELETE FROM choice WHERE id = ?
	`, choiceID)

	return err
}
