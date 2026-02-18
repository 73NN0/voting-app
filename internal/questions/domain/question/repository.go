package question

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	CreateQuestion(context.Context, Question) (int /* question id */, error)
	GetQuestionByID(context.Context, int /*question id */) (Question, error)
	GetQuestionsBySessionID(context.Context, uuid.UUID /*session id */) ([]Question, error)
	DeleteQuestion(context.Context, int /*question id */) error
	UpdateQuestion(context.Context, Question) error
}
