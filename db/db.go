package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id    uuid.UUID
	Name  string
	Email string
	Mdp   string
}

func OpenDB(name, dns string) *sql.DB {

	db, err := sql.Open(name, dns)
	if err != nil {
		panic(fmt.Sprintf("failed to open db: %v", err))
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		panic(fmt.Sprintf("err cannot ping database server err is %q", err))
	}

	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		panic(fmt.Sprintf("failed to enable foreign keys: %v", err))
	}

	return db
}

func CreateUser(ctx context.Context, db *sql.DB, user User) error {

	_, err := db.ExecContext(ctx, `INSERT INTO user (id, Name, email) VALUES(?, ?, ?);`, user.Id.String(), user.Name, user.Email)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `INSERT INTO user_password (user_id, password_hash ) VALUES(?, ?);`, user.Id.String(), user.Mdp)

	return err
}

func GetUserByID(ctx context.Context, db *sql.DB, id string) (*User, error) {
	if id == "" {
		return nil, fmt.Errorf("invalid id")
	}
	user := User{}

	if err := db.QueryRowContext(ctx, `SELECT id, name, email FROM user WHERE id = ?;`, id).Scan(&user.Id, &user.Name, &user.Email); err != nil {
		return nil, err
	}

	if err := db.QueryRowContext(ctx, `SELECT password_hash FROM user_password WHERE user_id = ?;`, id).Scan(&user.Mdp); err != nil {
		return nil, err
	}

	return &user, nil
}

func GetUserByEmail(ctx context.Context, db *sql.DB, email string) (*User, error) {
	if email == "" {
		return nil, fmt.Errorf("invalid email")
	}

	user := User{}

	if err := db.QueryRowContext(ctx, `SELECT id, name, email FROM user WHERE email = ?;`, email).Scan(&user.Id, &user.Name, &user.Email); err != nil {
		return nil, err
	}

	if err := db.QueryRowContext(ctx, `SELECT password_hash FROM user_password WHERE user_id = ?;`, user.Id.String()).Scan(&user.Mdp); err != nil {
		return nil, err
	}

	return &user, nil

}

func UserExists(ctx context.Context, db *sql.DB, id string) (bool, error) {
	var exists bool
	err := db.QueryRowContext(ctx, `
        SELECT EXISTS(SELECT 1 FROM user WHERE id = ?)
    `, id).Scan(&exists)

	return exists, err
}

func UpdateUser(ctx context.Context, db *sql.DB, user User) error {
	id := user.Id.String()
	if exist, err := UserExists(ctx, db, id); err != nil || !exist {
		return fmt.Errorf("the user may not exist (%v), abort, err is %v", exist, err)
	}

	stmt := `
UPDATE user
SET 
    name = ?,
    email = ?
WHERE id = ?
	`
	_, err := db.ExecContext(ctx, stmt, user.Name, user.Email, user.Id)

	// note: separate password.

	return err
}

// TODO later user COMMIT TRANSACTION TO SECURE MORE THE DELETION ( WITH ROLLBACK !)
func DeleteUser(ctx context.Context, db *sql.DB, user User) error {
	id := user.Id.String()

	stmt := `DELETE FROM user WHERE id = ?;`

	_, err := db.ExecContext(ctx, stmt, id)

	return err
}

type Session struct {
	Id          uuid.UUID
	Title       string
	Description string
	CreatedAt   Timestamp
	EndsAt      Timestamp
}

// ============= session

func CreateVoteSession(ctx context.Context, db *sql.DB, session Session) error {

	idString := session.Id.String()

	// Si CreatedAt est zero, utiliser DEFAULT
	if session.CreatedAt.IsZero() {
		_, err := db.ExecContext(ctx, `
            INSERT INTO vote_session (id, title, description, ends_at) 
            VALUES (?, ?, ?, ?)
        `, idString, session.Title, session.Description, session.EndsAt)
		return err
	}

	// Sinon, insérer explicitement
	_, err := db.ExecContext(ctx, `
        INSERT INTO vote_session (id, title, description, created_at, ends_at) 
        VALUES (?, ?, ?, ?, ?)
    `, idString, session.Title, session.Description, session.CreatedAt, session.EndsAt)
	return err
}

func GetVoteSessionByID(ctx context.Context, db *sql.DB, id string) (*Session, error) {
	var session Session

	err := db.QueryRowContext(ctx, `
        SELECT id, title, description, created_at, ends_at
        FROM vote_session WHERE id = ?
    `, id).Scan(&session.Id, &session.Title, &session.Description, &session.CreatedAt, &session.EndsAt)

	if err != nil {
		return nil, err
	}

	return &session, nil
}

// GIVEN: A user ID
// WHEN: Fetching all sessions the user participates in
// THEN: Returns list of sessions with their details
func GetUserVoteSessions(ctx context.Context, db *sql.DB, userID string) ([]*Session, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT vs.id, vs.title, vs.description, vs.created_at, vs.ends_at
		FROM vote_session vs
		INNER JOIN session_and_participant sp ON vs.id = sp.session_id
		WHERE sp.user_id = ?
		ORDER BY vs.created_at DESC
	`, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		session := &Session{}
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
		sessions = append(sessions, session)
	}

	return sessions, rows.Err()
}

func UpdateVoteSession(ctx context.Context, db *sql.DB, session Session) error {
	_, err := db.ExecContext(ctx, `
		UPDATE vote_session
		SET title = ?, description = ?, ends_at = ?
		WHERE id = ?
	`, session.Title, session.Description, session.EndsAt, session.Id.String())

	return err
}

func DeleteVoteSession(ctx context.Context, db *sql.DB, sessionID string) error {
	_, err := db.ExecContext(ctx, `
		DELETE FROM vote_session WHERE id = ?
	`, sessionID)

	return err
}

func CloseVoteSession(ctx context.Context, db *sql.DB, sessionID string) error {
	_, err := db.ExecContext(ctx, `
		UPDATE vote_session
		SET ends_at = ?
		WHERE id = ?
	`, Timestamp{time.Now().UTC()}, sessionID)

	return err
}

func ListVoteSessions(ctx context.Context, db *sql.DB, limit, offset int) ([]*Session, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, title, description, created_at, ends_at
		FROM vote_session
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, limit, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		session := &Session{}
		err := rows.Scan(&session.Id, &session.Title, &session.Description, &session.CreatedAt, &session.EndsAt)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	return sessions, rows.Err()
}

// session_and_participants

func AddParticipant(ctx context.Context, db *sql.DB, sessionID, userID string) error {
	_, err := db.ExecContext(ctx, `
        INSERT INTO session_and_participant (user_id, session_id, invited_at)
        VALUES (?, ?, ?)
    `, userID, sessionID, Timestamp{time.Now()})

	return err
}

func GetParticipants(ctx context.Context, db *sql.DB, sessionID string) ([]*User, error) {
	rows, err := db.QueryContext(ctx, `
        SELECT u.id, u.name, u.email
        FROM user u
        INNER JOIN session_and_participant sp ON u.id = sp.user_id
        WHERE sp.session_id = ?
    `, sessionID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		err := rows.Scan(&user.Id, &user.Name, &user.Email)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

// GIVEN: A session ID and user ID
// WHEN: Removing user from session participants
// THEN: User is no longer a participant
func RemoveParticipant(ctx context.Context, db *sql.DB, sessionID, userID string) error {
	_, err := db.ExecContext(ctx, `
		DELETE FROM session_and_participant
		WHERE session_id = ? AND user_id = ?
	`, sessionID, userID)

	return err
}

func IsParticipant(ctx context.Context, db *sql.DB, sessionID, userID string) (bool, error) {
	var exists bool
	err := db.QueryRowContext(ctx, `
        SELECT EXISTS(
            SELECT 1 FROM session_and_participant 
            WHERE session_id = ? AND user_id = ?
        )
    `, sessionID, userID).Scan(&exists)

	return exists, err
}

// ==== choices

type Question struct {
	ID            int // AUTOINCREMENT
	SessionID     string
	Text          string
	OrderNum      int
	AllowMultiple bool
	MaxChoices    int
	CreatedAt     Timestamp
}

type Choice struct {
	ID         int // AUTOINCREMENT
	QuestionID int
	Text       string
	OrderNum   int
	CreatedAt  Timestamp
}

// ============= Question =============

func CreateQuestion(ctx context.Context, db *sql.DB, question Question) (questionID int, err error) {
	result, err := db.ExecContext(ctx, `
		INSERT INTO question (session_id, text, order_num, allow_multiple, max_choices)
		VALUES (?, ?, ?, ?, ?)
	`, question.SessionID, question.Text, question.OrderNum, question.AllowMultiple, question.MaxChoices)

	if err != nil {
		return
	}

	id, err := result.LastInsertId()
	questionID = int(id)

	return
}

func GetQuestions(ctx context.Context, db *sql.DB, sessionID string) ([]*Question, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, session_id, text, order_num, allow_multiple, max_choices, created_at
		FROM question
		WHERE session_id = ?
		ORDER BY order_num ASC
	`, sessionID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []*Question
	for rows.Next() {
		q := &Question{}
		err := rows.Scan(&q.ID, &q.SessionID, &q.Text, &q.OrderNum, &q.AllowMultiple, &q.MaxChoices, &q.CreatedAt)
		if err != nil {
			return nil, err
		}
		questions = append(questions, q)
	}

	return questions, rows.Err()
}

func GetQuestionByID(ctx context.Context, db *sql.DB, id int) (*Question, error) {
	q := &Question{}

	err := db.QueryRowContext(ctx, `
		SELECT id, session_id, text, order_num, allow_multiple, max_choices, created_at
		FROM question
		WHERE id = ?
	`, id).Scan(&q.ID, &q.SessionID, &q.Text, &q.OrderNum, &q.AllowMultiple, &q.MaxChoices, &q.CreatedAt)

	if err != nil {
		return nil, err
	}

	return q, nil
}

func DeleteQuestion(ctx context.Context, db *sql.DB, id int) error {
	_, err := db.ExecContext(ctx, `
		DELETE FROM question WHERE id = ?
	`, id)

	return err
}

// ============= Choice =============

func CreateChoice(ctx context.Context, db *sql.DB, choice Choice) (choiceID int, err error) {
	result, err := db.ExecContext(ctx, `
		INSERT INTO choice (question_id, text, order_num)
		VALUES (?, ?, ?)
	`, choice.QuestionID, choice.Text, choice.OrderNum)

	if err != nil {
		return
	}
	id, err := result.LastInsertId()
	choiceID = int(id)

	return
}

func GetChoices(ctx context.Context, db *sql.DB, questionID int) ([]*Choice, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, question_id, text, order_num, created_at
		FROM choice
		WHERE question_id = ?
		ORDER BY order_num ASC
	`, questionID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var choices []*Choice
	for rows.Next() {
		c := &Choice{}
		err := rows.Scan(&c.ID, &c.QuestionID, &c.Text, &c.OrderNum, &c.CreatedAt)
		if err != nil {
			return nil, err
		}
		choices = append(choices, c)
	}

	return choices, rows.Err()
}

func GetChoiceByID(ctx context.Context, db *sql.DB, id int) (*Choice, error) {
	c := &Choice{}

	err := db.QueryRowContext(ctx, `
		SELECT id, question_id, text, order_num, created_at
		FROM choice
		WHERE id = ?
	`, id).Scan(&c.ID, &c.QuestionID, &c.Text, &c.OrderNum, &c.CreatedAt)

	if err != nil {
		return nil, err
	}

	return c, nil
}

func DeleteChoice(ctx context.Context, db *sql.DB, id int) error {
	_, err := db.ExecContext(ctx, `
		DELETE FROM choice WHERE id = ?
	`, id)

	return err
}
