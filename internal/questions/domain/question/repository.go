package questions

import (
	"context"

	"github.com/google/uuid"
)

type Respository interface {
	CreateQuestion(context.Context, Question) (int /* question id */, error)
	GetQuestionByID(context.Context /*question id */, int) (*Question, error)
	GetQuestionsBySessionID(context.Context, uuid.UUID /*session id */) ([]*Question, error)
	DeleteQuestion(context.Context /*question id */, int)
}
