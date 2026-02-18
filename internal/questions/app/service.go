package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/73NN0/voting-app/internal/questions/domain/choice"
	"github.com/73NN0/voting-app/internal/questions/domain/question"
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
