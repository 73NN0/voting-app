package adapters

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/73NN0/voting-app/internal/common/db"
	"github.com/73NN0/voting-app/internal/users/domain/user"
	"github.com/google/uuid"
)

// ========== DTOs ==========

// userDTO représente la table "user"
type userDTO struct {
	ID        string       // TEXT (uuid)
	Name      string       // TEXT
	Email     string       // TEXT
	CreatedAt db.Timestamp // TEXT
}

// userPasswordDTO représente la table "user_password"
type userPasswordDTO struct {
	UserID       string       // TEXT
	PasswordHash string       // TEXT
	CreatedAt    db.Timestamp // TEXT
	UpdatedAt    db.Timestamp // TEXT
}

// ========== Conversions Domain → DTO ==========

func toUserDTO(u *user.User) userDTO {
	return userDTO{
		ID:        u.ID().String(),
		Name:      u.Name(),
		Email:     u.Email(),
		CreatedAt: db.Timestamp{Time: u.CreatedAt()},
	}
}

// ========== Conversions DTO → Domain ==========

func (dto userDTO) toDomain() (*user.User, error) {
	id, err := uuid.Parse(dto.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	return user.UnmarshalUserFromRepository(
		id,
		dto.Name,
		dto.Email,
		dto.CreatedAt.Time,
	)
}

// ========== Repository Implementation ==========

type SqliteUserRepository struct {
	db db.DBRepository
}

// Compile-time check
var _ user.Repository = (*SqliteUserRepository)(nil)

func NewSqliteUserRepository(database db.DBRepository) *SqliteUserRepository {
	return &SqliteUserRepository{db: database}
}

// ===== User CRUD =====

func (r *SqliteUserRepository) CreateUser(ctx context.Context, u *user.User) error {
	dto := toUserDTO(u)

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO "user" (id, name, email, created_at)
		VALUES (?, ?, ?, ?)
	`, dto.ID, dto.Name, dto.Email, dto.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	return nil
}

func (r *SqliteUserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	var dto userDTO

	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, email, created_at
		FROM "user"
		WHERE id = ?
	`, id.String()).Scan(&dto.ID, &dto.Name, &dto.Email, &dto.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user %s not found", id)
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	return dto.toDomain()
}

func (r *SqliteUserRepository) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {
	var dto userDTO

	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, email, created_at
		FROM "user"
		WHERE email = ?
	`, email).Scan(&dto.ID, &dto.Name, &dto.Email, &dto.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with email %s not found", email)
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	return dto.toDomain()
}

func (r *SqliteUserRepository) UpdateUser(ctx context.Context, u *user.User) error {
	dto := toUserDTO(u)

	_, err := r.db.ExecContext(ctx, `
		UPDATE "user"
		SET name = ?, email = ?
		WHERE id = ?
	`, dto.Name, dto.Email, dto.ID)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (r *SqliteUserRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM "user" WHERE id = ?`, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete user %s: %w", id, err)
	}
	return nil
}

func (r *SqliteUserRepository) ListUsers(ctx context.Context, limit, offset int) ([]*user.User, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, email, created_at
		FROM "user"
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*user.User
	for rows.Next() {
		var dto userDTO
		err := rows.Scan(&dto.ID, &dto.Name, &dto.Email, &dto.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		u, err := dto.toDomain()
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, rows.Err()
}

// ===== Password Management =====

func (r *SqliteUserRepository) SetPassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	// INSERT OR REPLACE pour gérer création + update
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO user_password (user_id, password_hash, created_at, updated_at)
		VALUES (?, ?, datetime('now'), datetime('now'))
		ON CONFLICT(user_id) DO UPDATE SET
			password_hash = excluded.password_hash,
			updated_at = datetime('now')
	`, userID.String(), passwordHash)

	if err != nil {
		return fmt.Errorf("failed to set password: %w", err)
	}

	return nil
}

func (r *SqliteUserRepository) GetPasswordHash(ctx context.Context, userID uuid.UUID) (string, error) {
	var passwordHash string

	err := r.db.QueryRowContext(ctx, `
		SELECT password_hash
		FROM user_password
		WHERE user_id = ?
	`, userID.String()).Scan(&passwordHash)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("password not found for user %s", userID)
		}
		return "", fmt.Errorf("failed to query password: %w", err)
	}

	return passwordHash, nil
}

func (r *SqliteUserRepository) DeletePassword(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM user_password WHERE user_id = ?
	`, userID.String())

	if err != nil {
		return fmt.Errorf("failed to delete password: %w", err)
	}

	return nil
}
