package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"gitlab.com/singfield/voting-app/entities"
)

type sessionDB struct {
	Id          string
	Title       string
	Description string
	CreatedAt   Timestamp
	EndsAt      Timestamp
}

type VoteSessionRepository struct {
	db *sql.DB
}

func NewVoteSessionRepository(db *sql.DB) *VoteSessionRepository {
	if db == nil {
		panic("no db")
	}

	return &VoteSessionRepository{
		db: db,
	}
}

func (s *sessionDB) toEntity() *entities.Session {

	id := stringToUuid(s.Id)

	return &entities.Session{
		Id:          id,
		Title:       s.Title,
		Description: s.Description,
		CreatedAt:   s.CreatedAt.Time,
		EndsAt:      s.EndsAt.Time,
	}
}

func fromEntitySession(s entities.Session) *sessionDB {

	id := uuidToString(s.Id)

	return &sessionDB{
		Id:          id,
		Title:       s.Title,
		Description: s.Description,
		CreatedAt:   Timestamp{s.CreatedAt.UTC()},
		EndsAt:      Timestamp{s.EndsAt.UTC()},
	}
}

func (v *VoteSessionRepository) CreateVoteSession(ctx context.Context, session entities.Session) error {

	// idString := uuidToString(session.Id)
	// note for a test : pass Time.Time to SQL... then get back the result...
	ss := fromEntitySession(session)

	// Si CreatedAt est zero, utiliser DEFAULT
	if session.CreatedAt.IsZero() {
		_, err := v.db.ExecContext(ctx, `
            INSERT INTO vote_session (id, title, description, ends_at) 
            VALUES (?, ?, ?, ?)
        `, ss.Id, ss.Title, ss.Description, ss.EndsAt)
		return err
	}

	// Sinon, insérer explicitement
	_, err := v.db.ExecContext(ctx, `
        INSERT INTO vote_session (id, title, description, created_at, ends_at) 
        VALUES (?, ?, ?, ?, ?)
    `, ss.Id, ss.Title, ss.Description, ss.CreatedAt, ss.EndsAt)
	return err
}

func (v *VoteSessionRepository) GetVoteSessionByID(ctx context.Context, id uuid.UUID) (*entities.Session, error) {
	var session sessionDB

	err := v.db.QueryRowContext(ctx, `
        SELECT id, title, description, created_at, ends_at
        FROM vote_session WHERE id = ?
    `, id).Scan(&session.Id, &session.Title, &session.Description, &session.CreatedAt, &session.EndsAt)

	if err != nil {
		return nil, err
	}

	return session.toEntity(), nil
}

// GIVEN: A user ID
// WHEN: Fetching all sessions the user participates in
// THEN: Returns list of sessions with their details
func (v *VoteSessionRepository) GetUserVoteSessions(ctx context.Context, userID uuid.UUID) ([]*entities.Session, error) {
	rows, err := v.db.QueryContext(ctx, `
		SELECT vs.id, vs.title, vs.description, vs.created_at, vs.ends_at
		FROM vote_session vs
		INNER JOIN session_and_participant sp ON vs.id = sp.session_id
		WHERE sp.user_id = ?
		ORDER BY vs.created_at DESC
	`, uuidToString(userID))

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*entities.Session
	for rows.Next() {
		session := &sessionDB{}
		err := rows.Scan(
			&session.Id,
			&session.Title,
			&session.Description,
			&session.CreatedAt,
			&session.EndsAt,
		)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session.toEntity())
	}

	return sessions, rows.Err()
}

func (v *VoteSessionRepository) UpdateVoteSession(ctx context.Context, session entities.Session) error {
	s := fromEntitySession(session)

	_, err := v.db.ExecContext(ctx, `
		UPDATE vote_session
		SET title = ?, description = ?, ends_at = ?
		WHERE id = ?
	`, s.Title, s.Description, s.EndsAt, s.Id)

	return err
}

func (v *VoteSessionRepository) DeleteVoteSession(ctx context.Context, sessionID uuid.UUID) error {
	id := uuidToString(sessionID)

	_, err := v.db.ExecContext(ctx, `
		DELETE FROM vote_session WHERE id = ?
	`, id)

	return err
}

func (v *VoteSessionRepository) CloseVoteSession(ctx context.Context, sessionID uuid.UUID) error {
	id := uuidToString(sessionID)
	_, err := v.db.ExecContext(ctx, `
		UPDATE vote_session
		SET ends_at = ?
		WHERE id = ?
	`, Timestamp{time.Now().UTC()}, id)

	return err
}

func (v *VoteSessionRepository) ListVoteSessions(ctx context.Context, limit, offset int) ([]*entities.Session, error) {

	rows, err := v.db.QueryContext(ctx, `
		SELECT id, title, description, created_at, ends_at
		FROM vote_session
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, limit, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*entities.Session
	for rows.Next() {
		session := &sessionDB{}
		err := rows.Scan(&session.Id, &session.Title, &session.Description, &session.CreatedAt, &session.EndsAt)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session.toEntity())
	}

	return sessions, rows.Err()
}

func (v *VoteSessionRepository) AddParticipant(ctx context.Context, sessionID, userID uuid.UUID) error {
	sID := uuidToString(sessionID)
	uID := uuidToString(userID)
	t := Timestamp{time.Now()}

	_, err := v.db.ExecContext(ctx, `
        INSERT INTO session_and_participant (user_id, session_id, invited_at)
        VALUES (?, ?, ?)
    `, uID, sID, t)

	return err
}

func (v *VoteSessionRepository) GetParticipants(ctx context.Context, sessionID uuid.UUID) ([]*entities.User, error) {

	sID := uuidToString(sessionID)

	rows, err := v.db.QueryContext(ctx, `
        SELECT u.id, u.name, u.email
        FROM user u
        INNER JOIN session_and_participant sp ON u.id = sp.user_id
        WHERE sp.session_id = ?
    `, sID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*entities.User
	for rows.Next() {
		user := &userDB{}
		err := rows.Scan(&user.Id, &user.Name, &user.Email)
		if err != nil {
			return nil, err
		}
		users = append(users, user.toEntity())
	}

	return users, rows.Err()
}

// GIVEN: A session ID and user ID
// WHEN: Removing user from session participants
// THEN: User is no longer a participant
func (v *VoteSessionRepository) RemoveParticipant(ctx context.Context, sessionID, userID uuid.UUID) error {
	_, err := v.db.ExecContext(ctx, `
		DELETE FROM session_and_participant
		WHERE session_id = ? AND user_id = ?
	`, uuidToString(sessionID), uuidToString(userID))

	return err
}

func (v *VoteSessionRepository) IsParticipant(ctx context.Context, sessionID, userID uuid.UUID) (bool, error) {
	sID := uuidToString(sessionID)
	uID := uuidToString(userID)

	var exists bool

	err := v.db.QueryRowContext(ctx, `
        SELECT EXISTS(
            SELECT 1 FROM session_and_participant 
            WHERE session_id = ? AND user_id = ?
        )
    `, sID, uID).Scan(&exists)

	return exists, err
}
