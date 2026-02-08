package adapters

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/73NN0/voting-app/internal/common/db"
	"github.com/73NN0/voting-app/internal/sessions/domain/session"
	"github.com/73NN0/voting-app/internal/users/domain/user"
	"github.com/google/uuid"
)

// ========== DTOs ==========

// sessionDTO représente vote_session en DB
type sessionDTO struct {
	ID          string        // TEXT (uuid)
	Title       string        // TEXT
	Description string        // TEXT
	CreatedAt   db.Timestamp  // TEXT
	EndsAt      *db.Timestamp // TEXT nullable
}

// participantDTO représente session_and_participant en DB
type participantDTO struct {
	UserID    string       // TEXT
	SessionID string       // TEXT
	InvitedAt db.Timestamp // TEXT
}

// ========== Conversions Domain → DTO ==========

func toSessionDTO(s *session.Session) sessionDTO {
	dto := sessionDTO{
		ID:          s.ID().String(),
		Title:       s.Title(),
		Description: s.Description(),
		CreatedAt:   db.Timestamp{Time: s.CreatedAt()},
	}

	// endsAt est optionnel
	if endsAt, ok := s.EndsAt(); ok {
		dto.EndsAt = &db.Timestamp{Time: endsAt}
	}

	return dto
}

// ========== Conversions DTO → Domain ==========

func (dto sessionDTO) toSession() (*session.Session, error) {
	id, err := uuid.Parse(dto.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid session id: %w", err)
	}

	// Unmarshal avec ou sans endsAt
	if dto.EndsAt != nil {
		return session.Rehydrate(
			id,
			dto.Title,
			dto.Description,
			dto.CreatedAt.Time,
			&dto.EndsAt.Time,
		)
	}

	return session.Rehydrate(
		id,
		dto.Title,
		dto.Description,
		dto.CreatedAt.Time,
		nil,
	)
}

// ========== Repository Implementation ==========

type SqliteSessionRepository struct {
	db db.DBRepository
}

// Compile-time check
var _ session.Repository = (*SqliteSessionRepository)(nil)

func NewSqliteSessionRepository(database db.DBRepository) *SqliteSessionRepository {
	return &SqliteSessionRepository{db: database}
}

// ===== Session CRUD =====

func (r *SqliteSessionRepository) CreateVoteSession(ctx context.Context, s *session.Session) error {
	dto := toSessionDTO(s)

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO vote_session (id, title, description, created_at, ends_at)
		VALUES (?, ?, ?, ?, ?)
	`, dto.ID, dto.Title, dto.Description, dto.CreatedAt, dto.EndsAt)

	if err != nil {
		return fmt.Errorf("failed to insert session: %w", err)
	}

	return nil
}

func (r *SqliteSessionRepository) GetVoteSessionByID(ctx context.Context, id uuid.UUID) (*session.Session, error) {
	var dto sessionDTO

	err := r.db.QueryRowContext(ctx, `
		SELECT id, title, description, created_at, ends_at
		FROM vote_session
		WHERE id = ?
	`, id.String()).Scan(
		&dto.ID,
		&dto.Title,
		&dto.Description,
		&dto.CreatedAt,
		&dto.EndsAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session %s not found", id)
		}
		return nil, fmt.Errorf("failed to query session: %w", err)
	}

	return dto.toSession()
}

func (r *SqliteSessionRepository) GetUserVoteSessions(ctx context.Context, userID uuid.UUID) ([]*session.Session, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT vs.id, vs.title, vs.description, vs.created_at, vs.ends_at
		FROM vote_session vs
		INNER JOIN session_and_participant sp ON vs.id = sp.session_id
		WHERE sp.user_id = ?
		ORDER BY vs.created_at DESC
	`, userID.String())

	if err != nil {
		return nil, fmt.Errorf("failed to query user sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*session.Session
	for rows.Next() {
		var dto sessionDTO
		err := rows.Scan(&dto.ID, &dto.Title, &dto.Description, &dto.CreatedAt, &dto.EndsAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}

		s, err := dto.toSession()
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}

	return sessions, rows.Err()
}

func (r *SqliteSessionRepository) UpdateVoteSession(ctx context.Context, s *session.Session) error {
	dto := toSessionDTO(s)

	_, err := r.db.ExecContext(ctx, `
		UPDATE vote_session
		SET title = ?, description = ?, ends_at = ?
		WHERE id = ?
	`, dto.Title, dto.Description, dto.EndsAt, dto.ID)

	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

func (r *SqliteSessionRepository) DeleteVoteSession(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM vote_session WHERE id = ?`, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete session %s: %w", id, err)
	}
	return nil
}

func (r *SqliteSessionRepository) CloseVoteSession(ctx context.Context, id uuid.UUID) error {
	now := db.Timestamp{Time: time.Now().UTC()}

	_, err := r.db.ExecContext(ctx, `
		UPDATE vote_session
		SET ends_at = ?
		WHERE id = ?
	`, now, id.String())

	if err != nil {
		return fmt.Errorf("failed to close session: %w", err)
	}

	return nil
}

func (r *SqliteSessionRepository) ListVoteSessions(ctx context.Context, limit, offset int) ([]*session.Session, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, description, created_at, ends_at
		FROM vote_session
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*session.Session
	for rows.Next() {
		var dto sessionDTO
		err := rows.Scan(&dto.ID, &dto.Title, &dto.Description, &dto.CreatedAt, &dto.EndsAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}

		s, err := dto.toSession()
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}

	return sessions, rows.Err()
}

// ===== Participants =====

func (r *SqliteSessionRepository) AddParticipant(ctx context.Context, sessionID, userID uuid.UUID) error {
	now := db.Timestamp{Time: time.Now().UTC()}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO session_and_participant (user_id, session_id, invited_at)
		VALUES (?, ?, ?)
	`, userID.String(), sessionID.String(), now)

	if err != nil {
		return fmt.Errorf("failed to add participant: %w", err)
	}

	return nil
}

func (r *SqliteSessionRepository) GetParticipants(ctx context.Context, sessionID uuid.UUID) ([]*user.User, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT u.id, u.name, u.email, u.created_at
		FROM "user" u
		INNER JOIN session_and_participant sp ON u.id = sp.user_id
		WHERE sp.session_id = ?
		ORDER BY sp.invited_at ASC
	`, sessionID.String())

	if err != nil {
		return nil, fmt.Errorf("failed to query participants: %w", err)
	}
	defer rows.Close()

	var users []*user.User
	for rows.Next() {
		var id, name, email string
		var createdAt db.Timestamp

		err := rows.Scan(&id, &name, &email, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan participant: %w", err)
		}

		userID, _ := uuid.Parse(id)

		// TODO: Adapter selon ton domain user.User
		// Pour l'instant je crée directement (à ajuster)
		u, err := user.Rehydrate(userID, name, email, createdAt.Time)
		if err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	return users, rows.Err()
}

func (r *SqliteSessionRepository) RemoveParticipant(ctx context.Context, sessionID, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM session_and_participant
		WHERE session_id = ? AND user_id = ?
	`, sessionID.String(), userID.String())

	if err != nil {
		return fmt.Errorf("failed to remove participant: %w", err)
	}

	return nil
}

func (r *SqliteSessionRepository) IsParticipant(ctx context.Context, sessionID, userID uuid.UUID) (bool, error) {
	var count int

	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM session_and_participant
		WHERE session_id = ? AND user_id = ?
	`, sessionID.String(), userID.String()).Scan(&count)

	if err != nil {
		return false, fmt.Errorf("failed to check participant: %w", err)
	}

	return count > 0, nil
}
