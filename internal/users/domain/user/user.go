package user

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
)

type User struct {
	id        uuid.UUID
	name      string
	email     string
	createdAt time.Time
}

var (
	ErrEmptyName     = errors.New("user name cannot be empty")
	ErrInvalidEmail  = errors.New("invalid email format")
	ErrInvalidUserID = errors.New("invalid user id")
)

// Regex simple pour validation email
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// Getters read-only
func (u *User) ID() uuid.UUID        { return u.id }
func (u *User) Name() string         { return u.name }
func (u *User) Email() string        { return u.email }
func (u *User) CreatedAt() time.Time { return u.createdAt }

// Constructeur
func NewUser(name, email string) (*User, error) {
	if name == "" {
		return nil, ErrEmptyName
	}

	if !emailRegex.MatchString(email) {
		return nil, fmt.Errorf("%w: %s", ErrInvalidEmail, email)
	}

	return &User{
		id:        uuid.New(),
		name:      name,
		email:     email,
		createdAt: time.Now().UTC(),
	}, nil
}

func (u *User) UpdateName(newName string) error {
	if newName == "" {
		return ErrEmptyName
	}
	u.name = newName
	return nil
}

func (u *User) UpdateEmail(newEmail string) error {
	if !emailRegex.MatchString(newEmail) {
		return fmt.Errorf("%w: %s", ErrInvalidEmail, newEmail)
	}
	u.email = newEmail
	return nil
}

func Rehydrate(id uuid.UUID, name, email string, createdAt time.Time) (*User, error) {
	if id == uuid.Nil {
		return nil, ErrInvalidUserID
	}

	if name == "" {
		return nil, ErrEmptyName
	}
	// Note : don't revalidate email. trust db
	return &User{
		id:        id,
		name:      name,
		email:     email,
		createdAt: createdAt,
	}, nil
}
