package adapters

import (
	"context"
	"errors"

	"github.com/73NN0/voting-app/internal/questions/app"
	"github.com/73NN0/voting-app/internal/sessions/domain/session"
	"github.com/google/uuid"
)

type SessionCheckerInProcess struct {
	repo session.Repository
}

var _ app.SessionChecker = (*SessionCheckerInProcess)(nil)

func NewSessionCheckerInProcess(repo session.Repository) *SessionCheckerInProcess {
	if repo == nil {
		panic(" missing session repository")
	}

	return &SessionCheckerInProcess{repo: repo}
}

func (c *SessionCheckerInProcess) Exists(ctx context.Context, sessionID uuid.UUID) (bool, error) {
	_, err := c.repo.GetVoteSessionByID(ctx, sessionID)
	if err == nil {
		return true, nil
	}

	if errors.Is(err, session.ErrNotFound) {
		return false, nil
	}

	return false, err
}
