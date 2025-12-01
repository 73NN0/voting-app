package entities

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	Id          uuid.UUID
	Title       string
	Description string
	CreatedAt   time.Time
	EndsAt      time.Time
}
