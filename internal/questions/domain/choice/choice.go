package choice

import (
	"errors"
	"time"
)

type Choice struct {
	createdAt  time.Time
	text       string
	id         int // AUTOINCREMENT
	questionID int
	orderNum   int
}

var (
	ErrEmptyChoiceText    = errors.New("choice text cannot be empty")
	ErrInvalidChoiceOrder = errors.New("choice order_num must be >= 1")
)

// question: be able to modify by passing an optional argument ? TODO: see in the futur if I need to change the id of a project's entity struct
// note for the future : I don't like this, it remember be JAVA with getter and setters.
// note for the future: need to compare with handemade hero.
// TODO : note this reflexion somewhere else

// no pointer receiver : it's readonly !
func (c *Choice) ID() int             { return c.id }
func (c Choice) QuestionID() int      { return c.questionID }
func (c Choice) Text() string         { return c.text }
func (c Choice) OrderNum() int        { return c.orderNum }
func (c Choice) CreatedAt() time.Time { return c.createdAt }

func NewChoice(questionID, orderNum int, text string) Choice {
	return Choice{
		questionID: questionID,
		text:       text,
		orderNum:   orderNum,
	}
}

func NewChoiceWithID(id, questionID, orderNum int, text string) Choice {
	return Choice{
		id:         id,
		questionID: questionID,
		text:       text,
		orderNum:   orderNum,
	}
}

func UnmarshalChoiceFromRepository(
	id int,
	questionID int,
	text string,
	orderNum int,
	createdAt time.Time,
) (*Choice, error) {
	if id <= 0 {
		return nil, errors.New("invalid choice id")
	}
	if text == "" {
		return nil, ErrEmptyChoiceText
	}

	return &Choice{
		id:         id,
		questionID: questionID,
		text:       text,
		orderNum:   orderNum,
		createdAt:  createdAt,
	}, nil
}

func (c *Choice) ChangeOrderNum(newOrderNum int) error {
	if newOrderNum < 1 {
		return ErrInvalidChoiceOrder
	}
	c.orderNum = newOrderNum
	return nil
}

func (c *Choice) UpdateText(newText string) error {
	if newText == "" {
		return ErrEmptyChoiceText
	}
	c.text = newText
	return nil
}
