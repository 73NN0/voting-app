package questions

import (
	"context"
)

type ChoiceRepository interface {
	CreateChoice(context.Context, Choice) (int, error)
	GetChoices(context.Context, int /* question id */) ([]*Choice, error)
	GetChoiceByID(context.Context, int /* choice id */) (*Choice, error)
	DeleteChoice(context.Context, int /* choice id */) error
}
