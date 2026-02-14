package session

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	CreateVoteSession(context.Context, *Session) error

	GetVoteSessionByID(context.Context, uuid.UUID /*session id */) (*Session, error)

	GetUserVoteSessions(context.Context, uuid.UUID /*session id */) ([]*Session, error)

	UpdateVoteSession(context.Context, *Session) error

	DeleteVoteSession(context.Context, uuid.UUID /*session id */) error

	CloseVoteSession(context.Context, uuid.UUID /*session id */) error

	ListVoteSessions(context.Context, int /* limit */, int /*offset */) ([]*Session, error)

	AddParticipant(context.Context, uuid.UUID /*session id */, uuid.UUID /* user id */) error

	GetParticipants(context.Context, uuid.UUID /*session id */) (uuid.UUIDs /* user id */, error)

	RemoveParticipant(context.Context, uuid.UUID /*session id */, uuid.UUID /*user id */) error

	IsParticipant(context.Context, uuid.UUID /*session id */, uuid.UUID /*user id */) (bool, error)
}
