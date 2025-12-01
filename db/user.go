package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type User struct {
	Id    uuid.UUID
	Name  string
	Email string
	Mdp   string
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
