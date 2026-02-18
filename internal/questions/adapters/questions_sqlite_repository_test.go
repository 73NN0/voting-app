package adapters_test

import (
	"context"
	"errors"
	"testing"

	"github.com/73NN0/voting-app/internal/common/db"
	"github.com/73NN0/voting-app/internal/questions/adapters"
	"github.com/73NN0/voting-app/internal/questions/domain/question"
	"github.com/google/uuid"
)

// Helper safe : panique si erreur invalide (pour tests uniquement)
func mustNewQuestion(t *testing.T, sessionID uuid.UUID, text string, orderNum, maxChoices int, allowMultiple bool) question.Question {
	t.Helper()
	q, err := question.NewQuestion(sessionID, text, orderNum, maxChoices, allowMultiple)
	if err != nil {
		t.Fatalf("mustNewQuestion failed: %v", err)
	}
	return q
}

func TestCreateQuestion(t *testing.T) {

	database, cleanup, err := db.OpenSQLite(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	if err := db.InitializeSchemas(database); err != nil {
		t.Fatal(err)
	}

	repo := adapters.NewSqliteQuestionsRepository(database)
	ctx := context.Background()

	tests := []struct {
		name           string
		sessionID      uuid.UUID
		text           string
		orderNum       int
		maxChoices     int
		allowMultiple  bool
		wantErr        error // erreur attendue sur CreateQuestion
		wantIDPositive bool  // si création réussie, ID > 0 ?
	}{
		{
			name:           "happy path - question valide",
			sessionID:      uuid.New(),
			text:           "Quelle est ta couleur préférée ?",
			orderNum:       1,
			maxChoices:     1,
			allowMultiple:  false,
			wantErr:        nil,
			wantIDPositive: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := mustNewQuestion(t, tt.sessionID, tt.text, tt.orderNum, tt.maxChoices, tt.allowMultiple)

			id, err := repo.CreateQuestion(ctx, q)

			// Vérification erreur
			if tt.wantErr != nil {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			// Cas succès
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantIDPositive && id <= 0 {
				t.Error("expected positive ID, got", id)
			}

			// Vérification round-trip (optionnel mais très utile)
			if tt.wantErr == nil {
				fetched, err := repo.GetQuestionByID(ctx, id)
				if err != nil {
					t.Fatalf("GetQuestionByID failed: %v", err)
				}
				if fetched.Text() != tt.text {
					t.Errorf("expected text %q, got %q", tt.text, fetched.Text())
				}
				if fetched.OrderNum() != tt.orderNum {
					t.Errorf("expected orderNum %d, got %d", tt.orderNum, fetched.OrderNum())
				}
			}
		})
	}
}

// TODO : Créer deux questions même sessionID + même orderNum → erreur UNIQUE violation
// Vérifier que l’erreur est bien propagée (pas panic, pas erreur générique)
