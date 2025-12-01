package entities

import (
	"time"

	"github.com/google/uuid"
)

type Question struct {
	ID            int
	SessionID     uuid.UUID
	Text          string
	OrderNum      int
	AllowMultiple bool
	MaxChoices    int
	CreatedAt     time.Time
}
