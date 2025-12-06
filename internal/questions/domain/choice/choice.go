package questions

import "time"

type Choice struct {
	id         int // AUTOINCREMENT
	questionID int
	text       string
	orderNum   int
	createdAt  time.Time
}

// question: be able to modify by passing an optional argument ? TODO: see in the futur if I need to change the id of a project's entity struct
// note for the future : I don't like this, it remember be JAVA with getter and setters.
// note for the future: need to compare with handemade hero.
// TODO : note this reflexion somewhere else

// no pointer receiver : it's readonly !
func (c Choice) ID() int             { return c.id }
func (c Choice) QuestionID() int     { return c.questionID }
func (c Choice) Text() string        { return c.text }
func (c Choice) OrderNum() int       { return c.orderNum }
func (c Choice) CreateAt() time.Time { return c.createdAt }
