package session

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Session représente une session de vote
type Session struct {
	id          uuid.UUID
	title       string
	description string
	createdAt   time.Time
	endsAt      *time.Time // nullable
}

var (
	ErrNotFound         = errors.New("session not found")
	ErrEmptyTitle       = errors.New("session title cannot be empty")
	ErrInvalidSessionID = errors.New("invalid session id")
)

// Getters
func (s *Session) ID() uuid.UUID        { return s.id }
func (s *Session) Title() string        { return s.title }
func (s *Session) Description() string  { return s.description }
func (s *Session) CreatedAt() time.Time { return s.createdAt }

func (s *Session) HasEnd() bool {
	return s.endsAt != nil
}

func (s *Session) EndsAt() (time.Time, bool) {
	if s.endsAt == nil {
		return time.Time{}, false
	}
	return *s.endsAt, true
}

// Constructeurs
func NewSessionNoEnd(title, description string) (*Session, error) {
	if title == "" {
		return nil, ErrEmptyTitle
	}

	return &Session{
		id:          uuid.New(),
		title:       title,
		description: description,
		createdAt:   time.Now().UTC(),
		endsAt:      nil,
	}, nil
}

func NewSessionWithEnd(title, description string, endsAt time.Time) (*Session, error) {
	if title == "" {
		return nil, ErrEmptyTitle
	}

	return &Session{
		id:          uuid.New(),
		title:       title,
		description: description,
		createdAt:   time.Now().UTC(),
		endsAt:      &endsAt,
	}, nil
}

func Rehydrate(
	id uuid.UUID,
	title string,
	description string,
	createdAt time.Time,
	endsAt *time.Time,
) (*Session, error) {
	if id == uuid.Nil {
		return nil, ErrInvalidSessionID
	}
	if title == "" {
		return nil, ErrEmptyTitle
	}

	if endsAt != nil && endsAt.Before(createdAt) {
		return nil, errors.New("end date cannot be before creation date")
	}

	if description == "" && title != "" {
		description = title
	}

	return &Session{
		id:          id,
		title:       title,
		description: description,
		createdAt:   createdAt,
		endsAt:      endsAt,
	}, nil
}

// Comportement métier
func (s *Session) UpdateTitle(newTitle string) error {
	if newTitle == "" {
		return ErrEmptyTitle
	}
	s.title = newTitle
	return nil
}

func (s *Session) UpdateDescription(newDescription string) {
	s.description = newDescription
}

func (s *Session) SetEndDate(endsAt time.Time) {
	s.endsAt = &endsAt
}

func (s *Session) RemoveEndDate() {
	s.endsAt = nil
}
