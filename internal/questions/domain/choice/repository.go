package choice

import (
	"context"
)

type Repository interface {
	CreateChoice(context.Context, Choice) (int, error)
	GetChoiceByID(context.Context, int /* choice id */) (Choice, error)
	GetChoicesByQuestionID(context.Context, int /* question id */) ([]Choice, error)
	DeleteChoice(context.Context, int /* choice id */) error
	UpdateChoice(context.Context, Choice) error
}
