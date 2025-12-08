package user

import (
	"context"

	"github.com/google/uuid"
)

// Repository définit le contrat pour persister les users
type Repository interface {
	CreateUser(ctx context.Context, u *User) error
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateUser(ctx context.Context, u *User) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
	ListUsers(ctx context.Context, limit, offset int) ([]*User, error)

	// Password management (séparé du User)
	SetPassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
	GetPasswordHash(ctx context.Context, userID uuid.UUID) (string, error)
	DeletePassword(ctx context.Context, userID uuid.UUID) error
}
