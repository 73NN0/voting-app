package db_test

import (
	"testing"

	"github.com/google/uuid"
	"gitlab.com/singfield/voting-app/db"
	"gitlab.com/singfield/voting-app/entities"
)

func TestUser(t *testing.T) {
	dbConn, ctx, cleanup := setup(t, "sqlite", ":memory:")
	user := entities.User{
		Id:    uuid.New(),
		Name:  "anon",
		Email: "anon@anon.com",
		Mdp:   "1234",
	}

	defer cleanup()

	uRepository := db.NewUserRepository(dbConn)
	// GIVEN : an initialised SQL database
	// WHEN : calling CreateUser function
	// THEN: Create a new entry in the user table with clear name and clear email adress
	t.Run("create simple user", func(t *testing.T) {

		if err := uRepository.CreateUser(ctx, user); err != nil {
			t.Fatal(err)
		}
		var name, email, password_hash string

		if err := dbConn.QueryRowContext(ctx, "SELECT name, email FROM user WHERE id = ?;", user.Id.String()).Scan(&name, &email); err != nil {
			t.Fatal(err)
		}

		if name != user.Name || email != user.Email {
			t.Fatalf("user unknown got %s \n %s \n want %s \n %s", name, email, user.Id.String(), user.Email)
		}

		if err := dbConn.QueryRowContext(ctx, "SELECT password_hash FROM user_password WHERE user_id = ?;", user.Id.String()).Scan(&password_hash); err != nil {
			t.Fatal(err)
		}

		if password_hash != user.Mdp {
			t.Fatalf("no password got %s want %s", password_hash, user.Mdp)
		}
	})

	t.Run("get user by id", func(t *testing.T) {

		userQueried, err := uRepository.GetUserByID(ctx, user.Id)
		if err != nil {
			t.Fatal(err)
		}

		assertStructEqual(t, user, *userQueried)
	})

	t.Run("get user by email", func(t *testing.T) {
		userQueried, err := uRepository.GetUserByEmail(ctx, user.Email)
		if err != nil {
			t.Fatal(err)
		}

		assertStructEqual(t, user, *userQueried)
	})

	t.Run("update user", func(t *testing.T) {
		want := entities.User{
			Id:    user.Id,
			Email: "tenno@tenno.com",
			Name:  "tenno",
			Mdp:   user.Mdp,
		}
		err := uRepository.UpdateUser(ctx, want)
		if err != nil {
			t.Fatal(err)
		}

		userQueried, err := uRepository.GetUserByID(ctx, user.Id)
		if err != nil {
			t.Fatal(err)
		}

		assertStructEqual(t, want, *userQueried)

	})

	t.Run("delete user", func(t *testing.T) {
		err := uRepository.DeleteUser(ctx, user)
		if err != nil {
			t.Fatal(err)
		}

		if ok, err := uRepository.UserExists(ctx, user.Id); err != nil || ok {
			t.Fatalf("is the user destructed ? err : %v", err)
		}
	})
}
