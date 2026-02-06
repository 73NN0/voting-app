package adapters_test

import (
	"context"
	"testing"

	"github.com/73NN0/voting-app/internal/common/db"
	"github.com/73NN0/voting-app/internal/users/adapters"
	"github.com/73NN0/voting-app/internal/users/domain/user"
)

func setupTestDB(t *testing.T) (db.DBRepository, context.Context, func()) {
	t.Helper()

	database := db.NewSQLiteDBRepository()
	cleanup := database.OpenDB(":memory:")

	if err := database.InitializeDatabaseSchemas(); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	return database, ctx, cleanup
}

func TestUserRepository_CreateAndGet(t *testing.T) {
	database, ctx, cleanup := setupTestDB(t)
	defer cleanup()

	repo := adapters.NewSqliteUserRepository(database)

	// GIVEN: Un nouveau user
	u, err := user.NewUser("Alice", "alice@example.com")
	if err != nil {
		t.Fatal(err)
	}

	// WHEN: On le sauvegarde
	err = repo.CreateUser(ctx, u)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// THEN: On peut le récupérer
	fetched, err := repo.GetUserByID(ctx, u.ID())
	if err != nil {
		t.Fatalf("GetUserByID failed: %v", err)
	}

	// Vérifications
	if fetched.Name() != "Alice" {
		t.Errorf("name: got %q, want %q", fetched.Name(), "Alice")
	}

	if fetched.Email() != "alice@example.com" {
		t.Errorf("email: got %q, want %q", fetched.Email(), "alice@example.com")
	}

	if fetched.CreatedAt().IsZero() {
		t.Error("createdAt should not be zero")
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	database, ctx, cleanup := setupTestDB(t)
	defer cleanup()

	repo := adapters.NewSqliteUserRepository(database)

	// GIVEN: Un user
	u, _ := user.NewUser("Bob", "bob@example.com")
	repo.CreateUser(ctx, u)

	// WHEN: On cherche par email
	fetched, err := repo.GetUserByEmail(ctx, "bob@example.com")
	if err != nil {
		t.Fatal(err)
	}

	// THEN: On trouve le bon user
	if fetched.ID() != u.ID() {
		t.Error("should find user by email")
	}

	if fetched.Name() != "Bob" {
		t.Errorf("name: got %q, want Bob", fetched.Name())
	}
}

func TestUserRepository_Update(t *testing.T) {
	database, ctx, cleanup := setupTestDB(t)
	defer cleanup()

	repo := adapters.NewSqliteUserRepository(database)

	// GIVEN: Un user existant
	u, _ := user.NewUser("Alice", "alice@example.com")
	repo.CreateUser(ctx, u)

	// WHEN: On modifie ses infos
	u.UpdateName("Alice Smith")
	u.UpdateEmail("alice.smith@example.com")

	err := repo.UpdateUser(ctx, u)
	if err != nil {
		t.Fatal(err)
	}

	// THEN: Modifications persistées
	fetched, _ := repo.GetUserByID(ctx, u.ID())

	if fetched.Name() != "Alice Smith" {
		t.Error("name not updated")
	}

	if fetched.Email() != "alice.smith@example.com" {
		t.Error("email not updated")
	}
}

func TestUserRepository_Delete(t *testing.T) {
	database, ctx, cleanup := setupTestDB(t)
	defer cleanup()

	repo := adapters.NewSqliteUserRepository(database)

	// GIVEN: Un user
	u, _ := user.NewUser("Charlie", "charlie@example.com")
	repo.CreateUser(ctx, u)

	// WHEN: On le supprime
	err := repo.DeleteUser(ctx, u.ID())
	if err != nil {
		t.Fatal(err)
	}

	// THEN: Il n'existe plus
	_, err = repo.GetUserByID(ctx, u.ID())
	if err == nil {
		t.Error("expected error when fetching deleted user")
	}
}

func TestUserRepository_ListUsers(t *testing.T) {
	database, ctx, cleanup := setupTestDB(t)
	defer cleanup()

	repo := adapters.NewSqliteUserRepository(database)

	// GIVEN: Plusieurs adapters
	u1, _ := user.NewUser("Alice", "alice@example.com")
	u2, _ := user.NewUser("Bob", "bob@example.com")
	u3, _ := user.NewUser("Charlie", "charlie@example.com")

	repo.CreateUser(ctx, u1)
	repo.CreateUser(ctx, u2)
	repo.CreateUser(ctx, u3)

	// WHEN: On liste avec pagination
	adapters, err := repo.ListUsers(ctx, 2, 0) // Limite 2
	if err != nil {
		t.Fatal(err)
	}

	// THEN: 2 adapters
	if len(adapters) != 2 {
		t.Fatalf("expected 2 adapters, got %d", len(adapters))
	}

	// Test offset
	adapters2, _ := repo.ListUsers(ctx, 2, 1)
	if len(adapters2) != 2 {
		t.Fatalf("expected 2 adapters with offset, got %d", len(adapters2))
	}
}

func TestUserRepository_UniqueEmail(t *testing.T) {
	database, ctx, cleanup := setupTestDB(t)
	defer cleanup()

	repo := adapters.NewSqliteUserRepository(database)

	// GIVEN: Un user avec email
	u1, _ := user.NewUser("Alice", "alice@example.com")
	repo.CreateUser(ctx, u1)

	// WHEN: On essaie de créer un autre user avec le même email
	u2, _ := user.NewUser("Bob", "alice@example.com")
	err := repo.CreateUser(ctx, u2)

	// THEN: Erreur (UNIQUE constraint)
	if err == nil {
		t.Error("expected error for duplicate email")
	}
}

func TestUserRepository_SetPassword(t *testing.T) {
	database, ctx, cleanup := setupTestDB(t)
	defer cleanup()

	repo := adapters.NewSqliteUserRepository(database)

	// GIVEN: Un user
	u, _ := user.NewUser("Alice", "alice@example.com")
	repo.CreateUser(ctx, u)

	// WHEN: On set un password
	passwordHash := "$2a$10$hashedpassword"
	err := repo.SetPassword(ctx, u.ID(), passwordHash)
	if err != nil {
		t.Fatal(err)
	}

	// THEN: On peut le récupérer
	fetched, err := repo.GetPasswordHash(ctx, u.ID())
	if err != nil {
		t.Fatal(err)
	}

	if fetched != passwordHash {
		t.Errorf("password hash: got %q, want %q", fetched, passwordHash)
	}
}

func TestUserRepository_UpdatePassword(t *testing.T) {
	database, ctx, cleanup := setupTestDB(t)
	defer cleanup()

	repo := adapters.NewSqliteUserRepository(database)

	// GIVEN: User avec password
	u, _ := user.NewUser("Alice", "alice@example.com")
	repo.CreateUser(ctx, u)

	oldHash := "$2a$10$oldpassword"
	repo.SetPassword(ctx, u.ID(), oldHash)

	// WHEN: On update le password
	newHash := "$2a$10$newpassword"
	err := repo.SetPassword(ctx, u.ID(), newHash)
	if err != nil {
		t.Fatal(err)
	}

	// THEN: Nouveau password récupéré
	fetched, _ := repo.GetPasswordHash(ctx, u.ID())

	if fetched != newHash {
		t.Error("password should be updated")
	}

	if fetched == oldHash {
		t.Error("old password should be replaced")
	}
}

func TestUserRepository_DeletePassword(t *testing.T) {
	database, ctx, cleanup := setupTestDB(t)
	defer cleanup()

	repo := adapters.NewSqliteUserRepository(database)

	// GIVEN: User avec password
	u, _ := user.NewUser("Alice", "alice@example.com")
	repo.CreateUser(ctx, u)

	passwordHash := "$2a$10$hashedpassword"
	repo.SetPassword(ctx, u.ID(), passwordHash)

	// WHEN: On supprime le password
	err := repo.DeletePassword(ctx, u.ID())
	if err != nil {
		t.Fatal(err)
	}

	// THEN: Password n'existe plus
	_, err = repo.GetPasswordHash(ctx, u.ID())
	if err == nil {
		t.Error("expected error when fetching deleted password")
	}
}

func TestUserRepository_PasswordCascadeDelete(t *testing.T) {
	database, ctx, cleanup := setupTestDB(t)
	defer cleanup()

	repo := adapters.NewSqliteUserRepository(database)

	// GIVEN: User avec password
	u, _ := user.NewUser("Alice", "alice@example.com")
	repo.CreateUser(ctx, u)

	passwordHash := "$2a$10$hashedpassword"
	repo.SetPassword(ctx, u.ID(), passwordHash)

	// WHEN: On supprime le user
	err := repo.DeleteUser(ctx, u.ID())
	if err != nil {
		t.Fatal(err)
	}

	// THEN: Password CASCADE supprimé
	_, err = repo.GetPasswordHash(ctx, u.ID())
	if err == nil {
		t.Error("password should be cascade deleted with user")
	}
}

func TestUser_Validation(t *testing.T) {
	// Test que NewUser valide correctement

	// Nom vide
	_, err := user.NewUser("", "alice@example.com")
	if err != user.ErrEmptyName {
		t.Error("should reject empty name")
	}

	// Email invalide
	_, err = user.NewUser("Alice", "not-an-email")
	if err != user.ErrInvalidEmail {
		t.Error("should reject invalid email")
	}

	// Email valide
	u, err := user.NewUser("Alice", "alice@example.com")
	if err != nil {
		t.Error("should accept valid email")
	}

	if u == nil {
		t.Fatal("user should not be nil")
	}
}

func TestUser_UpdateValidation(t *testing.T) {
	u, _ := user.NewUser("Alice", "alice@example.com")

	// UpdateName avec vide
	err := u.UpdateName("")
	if err != user.ErrEmptyName {
		t.Error("should reject empty name")
	}

	// UpdateEmail avec invalide
	err = u.UpdateEmail("not-an-email")
	if err != user.ErrInvalidEmail {
		t.Error("should reject invalid email")
	}

	// UpdateEmail valide
	err = u.UpdateEmail("alice.new@example.com")
	if err != nil {
		t.Error("should accept valid email")
	}

	if u.Email() != "alice.new@example.com" {
		t.Error("email should be updated")
	}
}
