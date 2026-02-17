package adapters_test

import (
	"context"
	"testing"

	"github.com/73NN0/voting-app/internal/common/db"
	"github.com/73NN0/voting-app/internal/questions/adapters"
	"github.com/73NN0/voting-app/internal/questions/domain/question"
	"github.com/google/uuid"
)

func TestCreateQuestion(t *testing.T) {
	database, cleanup, err := db.OpenSQLite(":memory:")

	if err := db.InitializeSchemas(database); err != nil {
		t.Fatal(err)
	}

	defer cleanup()

	repo := adapters.NewSqliteQuestionsRepository(database)

	sessionID := uuid.New()

	q, err := question.NewQuestion(
		sessionID,
		"Quelle est ta couleur préférée ?",
		1,
		1,
		false,
	)

	if err != nil {
		t.Fatal(err)
	}

	id, err := repo.CreateQuestion(context.Background(), q)
	if err != nil {
		t.Fatalf("CreateQuestion failed: %v", err)
	}

	if id <= 0 {
		t.Error("expected positive question ID")
	}

	// Vérif read
	fetched, err := repo.GetQuestionByID(context.Background(), id)
	if err != nil {
		t.Fatalf("GetQuestionByID failed: %v", err)
	}
	if fetched.Text() != "Quelle est ta couleur préférée ?" {
		t.Errorf("expected text 'Quelle est ta couleur préférée ?', got '%s'", fetched.Text())
	}
}

// TODO : choice...
