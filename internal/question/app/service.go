package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/73NN0/voting-app/internal/question/domain/choice"
	"github.com/73NN0/voting-app/internal/question/domain/question"
	"github.com/google/uuid"
)

var (
	ErrVoteSessionNotFound = errors.New("vote session not found")
)

type SessionChecker interface {
	Exists(ctx context.Context, sessionID uuid.UUID) (bool, error)
}

type Service struct {
	questions question.Repository
	choices   choice.Repository
	sessions  SessionChecker
}

func NewService(questionRepository question.Repository, choiceRepository choice.Repository, sessions SessionChecker) *Service {
	if questionRepository == nil {
		panic("missing question repository")
	}

	if choiceRepository == nil {
		panic("missing choice repository")
	}

	if sessions == nil {
		panic("no Session access")
	}

	return &Service{
		questions: questionRepository,
		choices:   choiceRepository,
		sessions:  sessions,
	}
}

// questions

func (s *Service) CreateQuestion(ctx context.Context, sessionID uuid.UUID, text string, orderNum int, maxChoices int, allowMultiple bool) (int, error) {
	exists, err := s.sessions.Exists(ctx, sessionID)

	if err != nil {
		return 0, fmt.Errorf("check session: %w", err)
	}

	if !exists {
		return 0, ErrVoteSessionNotFound
	}

	q, err := question.NewQuestion(sessionID, text, orderNum, maxChoices, allowMultiple)

	if err != nil {
		return 0, err
	}

	return s.questions.CreateQuestion(ctx, q)
}

func (s *Service) GetQuestionByID(ctx context.Context, questionID int) (question.Question, error) {
	return s.questions.GetQuestionByID(ctx, questionID)
}

func (s *Service) ListQuestionsBySessionID(ctx context.Context, sessionID uuid.UUID) ([]question.Question, error) {
	return s.questions.GetQuestionsBySessionID(ctx, sessionID)
}

func (s *Service) DeleteQuestion(ctx context.Context, questionID int) error {
	return s.questions.DeleteQuestion(ctx, questionID)
}

func (s *Service) UpdateQuestion(ctx context.Context, id int, text string, orderNum, maxChoices int, allowMultiple bool) error {
	q, err := s.questions.GetQuestionByID(ctx, id)

	if err != nil {
		return err
	}

	if err := q.UpdateText(text); err != nil {
		return err
	}

	if err := q.ChangeOrderNum(orderNum); err != nil {
		return err
	}

	if q.AllowMultiple() != allowMultiple {
		q.ToggleAllowMultiple()
	}
	// TODO: maybe change to s.UpdateQuestion (the service fct become the interface)
	return s.questions.UpdateQuestion(ctx, q)
}

// choices

func (s *Service) CreateChoice(ctx context.Context, questionID int, orderNum int, text string) (int, error) {

	if questionID <= 0 {
		return 0, errors.New("invalid")
	}

	c := choice.NewChoice(questionID, orderNum, text)

	return s.choices.CreateChoice(ctx, c)
}

func (s *Service) ListChoicesByQuestionID(ctx context.Context, questionID int) ([]choice.Choice, error) {
	return s.choices.GetChoicesByQuestionID(ctx, questionID)
}

func (s *Service) DeleteChoice(ctx context.Context, choiceID int) error {
	return s.choices.DeleteChoice(ctx, choiceID)
}

// TODO updatedAt
func (s *Service) UpdateChoice(ctx context.Context, id int, text string, orderNum int) error {

	c, err := s.GetChoiceByID(ctx, id)

	if err != nil {
		return err
	}

	if err := c.UpdateText(text); err != nil {
		return err
	}
	if err := c.ChangeOrderNum(orderNum); err != nil {
		return err
	}

	return s.choices.UpdateChoice(ctx, c)
}

func (s *Service) GetChoiceByID(ctx context.Context, choiceID int) (choice.Choice, error) {
	return s.choices.GetChoiceByID(ctx, choiceID)
}

func (s *Service) ChangeChoiceQuestion(ctx context.Context, choiceID int, newQuestionID int) error {

	c, err := s.GetChoiceByID(ctx, choiceID)

	if err != nil {
		return err
	}

	exists, err := s.questions.IsQuestionExists(ctx, newQuestionID)

	if err != nil {
		return err
	}

	if !exists {
		return question.ErrNotFound
	}

	if err := c.UpdateQuestionID(newQuestionID); err != nil {
		return err
	}
	// TODO: maybe change to s.UpdateChoice (the service fct become the interface)
	return s.choices.UpdateChoice(ctx, c)
}
