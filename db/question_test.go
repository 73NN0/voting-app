package db_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"gitlab.com/singfield/voting-app/db"
	"gitlab.com/singfield/voting-app/entities"
)

func TestQuestion(t *testing.T) {
	dbConn, ctx, cleanup := setup(t, "sqlite", ":memory:")
	defer cleanup()
	sRepository := db.NewVoteSessionRepository(dbConn)
	// Setup: Create a session first
	session := entities.Session{
		Id:          uuid.New(),
		Title:       "AG 2025",
		Description: "Annual meeting",
		CreatedAt:   time.Now().UTC(),
		EndsAt:      time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC),
	}
	err := sRepository.CreateVoteSession(ctx, session)
	if err != nil {
		t.Fatal(err)
	}

	qRepository := db.NewQuestionRepository(dbConn)

	// GIVEN: A session exists
	// WHEN: Creating a question
	// THEN: Question is stored with correct order_num
	t.Run("create question", func(t *testing.T) {
		question := entities.Question{
			SessionID:     session.Id,
			Text:          "Approuvez-vous le budget 2025?",
			OrderNum:      1,
			AllowMultiple: false,
			MaxChoices:    1,
		}

		_, err := qRepository.CreateQuestion(ctx, question)
		if err != nil {
			t.Fatalf("failed to create question: %v", err)
		}

		// Verify question was created
		questions, err := qRepository.GetQuestions(ctx, session.Id)
		if err != nil {
			t.Fatal(err)
		}

		if len(questions) != 1 {
			t.Fatalf("expected 1 question, got %d", len(questions))
		}

		if questions[0].Text != question.Text {
			t.Errorf("text: got %q, want %q", questions[0].Text, question.Text)
		}
	})

	// GIVEN: A session with questions
	// WHEN: Creating multiple questions
	// THEN: Questions are ordered by order_num
	t.Run("create multiple questions with ordering", func(t *testing.T) {
		question2 := entities.Question{
			SessionID:     session.Id,
			Text:          "Élisez le nouveau président",
			OrderNum:      2,
			AllowMultiple: false,
			MaxChoices:    1,
		}

		question3 := entities.Question{
			SessionID:     session.Id,
			Text:          "Choix multiples test",
			OrderNum:      3,
			AllowMultiple: true,
			MaxChoices:    3,
		}

		_, err := qRepository.CreateQuestion(ctx, question2)
		if err != nil {
			t.Fatal(err)
		}

		_, err = qRepository.CreateQuestion(ctx, question3)
		if err != nil {
			t.Fatal(err)
		}

		// Get all questions
		questions, err := qRepository.GetQuestions(ctx, session.Id)
		if err != nil {
			t.Fatal(err)
		}

		if len(questions) != 3 {
			t.Fatalf("expected 3 questions, got %d", len(questions))
		}

		// Verify ordering
		for i := 0; i < len(questions)-1; i++ {
			if questions[i].OrderNum >= questions[i+1].OrderNum {
				t.Error("questions should be ordered by order_num ASC")
			}
		}
	})

	// GIVEN: A question exists
	// WHEN: Getting question by ID
	// THEN: Returns correct question
	t.Run("get question by id", func(t *testing.T) {
		questions, _ := qRepository.GetQuestions(ctx, session.Id)
		firstQuestion := questions[0]

		fetched, err := qRepository.GetQuestionByID(ctx, firstQuestion.ID)
		if err != nil {
			t.Fatalf("failed to get question: %v", err)
		}

		if fetched.Text != firstQuestion.Text {
			t.Errorf("text: got %q, want %q", fetched.Text, firstQuestion.Text)
		}
	})

	// GIVEN: A question with duplicate order_num
	// WHEN: Creating question
	// THEN: Should fail with UNIQUE constraint
	t.Run("duplicate order_num fails", func(t *testing.T) {
		duplicate := entities.Question{
			SessionID:     session.Id,
			Text:          "Duplicate order",
			OrderNum:      1, // Already exists
			AllowMultiple: false,
			MaxChoices:    1,
		}

		_, err := qRepository.CreateQuestion(ctx, duplicate)
		if err == nil {
			t.Error("expected error for duplicate order_num, got nil")
		}
	})

	// GIVEN: A session is deleted
	// WHEN: CASCADE delete triggers
	// THEN: Questions are deleted automatically
	t.Run("CASCADE delete session removes questions", func(t *testing.T) {
		// Create temp session with question
		tempSession := entities.Session{
			Id:          uuid.New(),
			Title:       "Temp",
			Description: "Will be deleted",
			CreatedAt:   time.Now().UTC(),
			EndsAt:      time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC),
		}
		sRepository.CreateVoteSession(ctx, tempSession)

		tempQuestion := entities.Question{
			SessionID:     tempSession.Id,
			Text:          "Temp question",
			OrderNum:      1,
			AllowMultiple: false,
			MaxChoices:    1,
		}
		qRepository.CreateQuestion(ctx, tempQuestion)

		// Delete session
		err := sRepository.DeleteVoteSession(ctx, tempSession.Id)
		if err != nil {
			t.Fatal(err)
		}

		// Verify questions were CASCADE deleted
		questions, err := qRepository.GetQuestions(ctx, tempSession.Id)
		if err != nil {
			t.Fatal(err)
		}

		if len(questions) != 0 {
			t.Errorf("expected 0 questions after CASCADE delete, got %d", len(questions))
		}
	})
}
