package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"gitlab.com/singfield/voting-app/entities"
)

type UserRepository struct {
	db *sql.DB
}

// TODO : Create_at & Updated_at
type userDB struct {
	Id    string
	Name  string
	Email string
	Mdp   string
}

func (u *userDB) toEntity() *entities.User {

	return &entities.User{
		Id:    stringToUuid(u.Id),
		Name:  u.Name,
		Email: u.Email,
		Mdp:   u.Mdp,
	}
}

func fromEntityUser(u entities.User) *userDB {

	return &userDB{
		Id:    uuidToString(u.Id),
		Name:  u.Name,
		Email: u.Email,
		Mdp:   u.Mdp,
	}
}

func NewUserRepository(db *sql.DB) *UserRepository {
	if db == nil {
		panic("no db")
	}

	return &UserRepository{
		db: db,
	}
}

func (u *UserRepository) CreateUser(ctx context.Context, user entities.User) error {

	_user := fromEntityUser(user)
	_, err := u.db.ExecContext(ctx, `INSERT INTO user (id, Name, email) VALUES(?, ?, ?);`, _user.Id, _user.Name, _user.Email)
	if err != nil {
		return err
	}
	_, err = u.db.ExecContext(ctx, `INSERT INTO user_password (user_id, password_hash ) VALUES(?, ?);`, _user.Id, _user.Mdp)

	return err
}

func (u *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {

	user := userDB{}
	idString := uuidToString(id)
	if err := u.db.QueryRowContext(ctx, `SELECT id, name, email FROM user WHERE id = ?;`, idString).Scan(&user.Id, &user.Name, &user.Email); err != nil {
		return nil, err
	}

	if err := u.db.QueryRowContext(ctx, `SELECT password_hash FROM user_password WHERE user_id = ?;`, idString).Scan(&user.Mdp); err != nil {
		return nil, err
	}

	return user.toEntity(), nil
}

func (u *UserRepository) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	if email == "" {
		return nil, fmt.Errorf("invalid email")
	}

	user := userDB{}

	if err := u.db.QueryRowContext(ctx, `SELECT id, name, email FROM user WHERE email = ?;`, email).Scan(&user.Id, &user.Name, &user.Email); err != nil {
		return nil, err
	}

	if err := u.db.QueryRowContext(ctx, `SELECT password_hash FROM user_password WHERE user_id = ?;`, user.Id).Scan(&user.Mdp); err != nil {
		return nil, err
	}

	return user.toEntity(), nil

}

func (u *UserRepository) UserExists(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	err := u.db.QueryRowContext(ctx, `
        SELECT EXISTS(SELECT 1 FROM user WHERE id = ?)
    `, uuidToString(id)).Scan(&exists)

	return exists, err
}

func (u *UserRepository) UpdateUser(ctx context.Context, user entities.User) error {

	if exist, err := u.UserExists(ctx, user.Id); err != nil || !exist {
		return fmt.Errorf("the user may not exist (%v), abort, err is %v", exist, err)
	}

	_user := fromEntityUser(user)

	stmt := `
UPDATE user
SET 
    name = ?,
    email = ?
WHERE id = ?
	`
	_, err := u.db.ExecContext(ctx, stmt, _user.Name, _user.Email, _user.Id)

	// note: separate password.

	return err
}

// TODO later user COMMIT TRANSACTION TO SECURE MORE THE DELETION ( WITH ROLLBACK !)
func (u *UserRepository) DeleteUser(ctx context.Context, user entities.User) error {
	udb := fromEntityUser(user)

	stmt := `DELETE FROM user WHERE id = ?;`

	_, err := u.db.ExecContext(ctx, stmt, udb.Id)

	return err
}
