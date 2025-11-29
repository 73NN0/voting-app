package db_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"gitlab.com/singfield/voting-app/db"

	_ "modernc.org/sqlite"
)

//1. TestCreateUser
//2. TestGetUserByID
//3. TestGetUserByEmail
//4. TestUpdateUser
//5. TestDeleteUser

// note : refactor the test for hash mdp !
// note : need to hide password from simple query to no propagate it throught the app
// Test Commit / rollback before transaction --> later

func TestUser(t *testing.T) {
	dbConn, ctx, cleanup := setup(t, "sqlite", ":memory:")
	user := db.User{
		Id:    uuid.New(),
		Name:  "ayden",
		Email: "ayden@ayden.com",
		Mdp:   "1234",
	}

	defer cleanup()

	// GIVEN : an initialised SQL database
	// WHEN : calling CreateUser function
	// THEN: Create a new entry in the user table with clear name and clear email adress
	t.Run("create simple user", func(t *testing.T) {

		if err := db.CreateUser(ctx, dbConn, user); err != nil {
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

		userQueried, err := db.GetUserByID(ctx, dbConn, user.Id.String())
		if err != nil {
			t.Fatal(err)
		}

		assertUserEqual(t, user, *userQueried)
	})

	t.Run("get user by email", func(t *testing.T) {
		userQueried, err := db.GetUserByEmail(ctx, dbConn, user.Email)
		if err != nil {
			t.Fatal(err)
		}

		assertUserEqual(t, user, *userQueried)
	})

	t.Run("update user", func(t *testing.T) {
		want := db.User{
			Id:    user.Id,
			Email: "tenno@tenno.com",
			Name:  "tenno",
			Mdp:   user.Mdp,
		}
		err := db.UpdateUser(ctx, dbConn, want)
		if err != nil {
			t.Fatal(err)
		}

		userQueried, err := db.GetUserByID(ctx, dbConn, user.Id.String())
		if err != nil {
			t.Fatal(err)
		}

		assertUserEqual(t, want, *userQueried)

	})

	t.Run("delete user", func(t *testing.T) {
		err := db.DeleteUser(ctx, dbConn, user)
		if err != nil {
			t.Fatal(err)
		}

		if ok, err := db.UserExists(ctx, dbConn, user.Id.String()); err != nil || ok {
			t.Fatalf("is the user destructed ? err : %v", err)
		}
	})
}

func TestVoteSession(t *testing.T) {

	dbConn, ctx, cleanup := setup(t, "sqlite", ":memory:")
	defer cleanup()

	if err := db.CreateVoteSession(ctx, dbConn, "1111", "vote meeting", "this is a small description", time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
}

func assertUserEqual(t *testing.T, want, got db.User) {
	t.Helper()
	if !reflect.DeepEqual(want, got) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}
