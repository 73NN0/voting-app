package question_test

import (
	"errors"
	"testing"

	"github.com/73NN0/voting-app/internal/questions/domain/question"
	"github.com/google/uuid"
)

func TestNewQuestion_Validations(t *testing.T) {
	sessionID := uuid.New()

	tests := []struct {
		name          string
		text          string
		orderNum      int
		maxChoices    int
		allowMultiple bool
		wantErr       error
	}{
		{"texte vide", "", 1, 1, false, question.ErrEmptyText},
		{"order num 0", "ok", 0, 1, false, question.ErrInvalidOrderNum},
		{"max choices 0", "ok", 1, 0, false, question.ErrInvalidMaxChoice},
		{"happy", "Quelle couleur ?", 1, 1, false, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := question.NewQuestion(sessionID, tt.text, tt.orderNum, tt.maxChoices, tt.allowMultiple)
			if tt.wantErr != nil {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected %v, got %v", tt.wantErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
