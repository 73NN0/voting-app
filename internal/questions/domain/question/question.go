package questions

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Question struct {
	createdAt     time.Time
	sessionID     uuid.UUID
	text          string
	id            int
	orderNum      int
	maxChoices    int
	allowMultiple bool
}

var (
	ErrEmptyText        = errors.New("question text cannot be empty")
	ErrInvalidOrderNum  = errors.New("order_num must be >= 1")
	ErrInvalidMaxChoice = errors.New("max_choices must be >= 1")
)

func (q Question) ID() int              { return q.id }
func (q Question) SessionID() uuid.UUID { return q.sessionID }
func (q Question) Text() string         { return q.text }
func (q Question) OrderNum() int        { return q.orderNum }
func (q Question) AllowMultiple() bool  { return q.allowMultiple }
func (q Question) MaxChoices() int      { return q.maxChoices }
func (q Question) CreatedAt() time.Time { return q.createdAt }

func NewQuestion(sessionID uuid.UUID, text string, orderNum, maxChoices int, allowMultiple bool) (Question, error) {
	if text == "" {
		return Question{}, ErrEmptyText
	}
	if orderNum < 1 {
		return Question{}, ErrInvalidOrderNum
	}
	if maxChoices < 1 {
		return Question{}, ErrInvalidMaxChoice
	}

	return Question{
		// id is set by the database
		sessionID:     sessionID,
		text:          text,
		orderNum:      orderNum,
		maxChoices:    maxChoices,
		allowMultiple: allowMultiple,
		// create_at is set by the database
	}, nil
}

func MustNewQuestion(sessionID uuid.UUID, text string, orderNum, maxChoices int, allowMultiple bool) Question {
	question, err := NewQuestion(sessionID, text, orderNum, maxChoices, allowMultiple)
	if err != nil {
		panic(err)
	}

	return question
}

func (q *Question) UpdateText(newText string) error {
	if newText == "" {
		return ErrEmptyText
	}
	q.text = newText
	return nil
}

func (q *Question) ChangeOrderNum(newOrderNum int) error {
	if newOrderNum < 1 {
		return ErrInvalidOrderNum
	}
	q.orderNum = newOrderNum
	return nil
}

func UnmarshalQuestionFromRepository(
	id int,
	sessionID uuid.UUID,
	text string,
	orderNum int,
	allowMultiple bool,
	maxChoices int,
	createdAt time.Time,
) (*Question, error) {
	if id <= 0 {
		return nil, errors.New("invalid question id")
	}
	if text == "" {
		return nil, ErrEmptyText
	}

	return &Question{
		id:            id,
		sessionID:     sessionID,
		text:          text,
		orderNum:      orderNum,
		allowMultiple: allowMultiple,
		maxChoices:    maxChoices,
		createdAt:     createdAt,
	}, nil
}
