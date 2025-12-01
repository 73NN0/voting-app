package db_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"gitlab.com/singfield/voting-app/db"
)

func TestChoice(t *testing.T) {
	dbConn, ctx, cleanup := setup(t, "sqlite", ":memory:")
	defer cleanup()

	cRepository := db.NewChoiceRepository(dbConn)
	qRepository := db.NewQuestionRepository(dbConn)
	// Setup: Create session + question
	session := db.Session{
		Id:          uuid.New(),
		Title:       "Test Session",
		Description: "For choices",
		CreatedAt:   db.Timestamp{time.Now().UTC()},
		EndsAt:      db.Timestamp{time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC)},
	}
	db.CreateVoteSession(ctx, dbConn, session)

	question := db.Question{
		SessionID:     session.Id.String(),
		Text:          "Choose your favorite",
		OrderNum:      1,
		AllowMultiple: true,
		MaxChoices:    2,
	}
	qRepository.CreateQuestion(ctx, dbConn, question)

	// Get question ID
	questions, _ := qRepository.GetQuestions(ctx, dbConn, session.Id.String())
	questionID := questions[0].ID

	// GIVEN: A question exists
	// WHEN: Creating a choice
	// THEN: Choice is stored with correct order_num
	t.Run("create choice", func(t *testing.T) {
		choice := db.Choice{
			QuestionID: questionID,
			Text:       "Option A",
			OrderNum:   1,
		}

		_, err := cRepository.CreateChoice(ctx, dbConn, choice)
		if err != nil {
			t.Fatalf("failed to create choice: %v", err)
		}

		// Verify
		choices, err := cRepository.GetChoices(ctx, dbConn, questionID)
		if err != nil {
			t.Fatal(err)
		}

		if len(choices) != 1 {
			t.Fatalf("expected 1 choice, got %d", len(choices))
		}

		if choices[0].Text != choice.Text {
			t.Errorf("text: got %q, want %q", choices[0].Text, choice.Text)
		}
	})

	// GIVEN: A question with choices
	// WHEN: Creating multiple choices
	// THEN: Choices are ordered by order_num
	t.Run("create multiple choices with ordering", func(t *testing.T) {
		choice2 := db.Choice{
			QuestionID: questionID,
			Text:       "Option B",
			OrderNum:   2,
		}

		choice3 := db.Choice{
			QuestionID: questionID,
			Text:       "Option C",
			OrderNum:   3,
		}

		cRepository.CreateChoice(ctx, dbConn, choice2)
		cRepository.CreateChoice(ctx, dbConn, choice3)

		choices, err := cRepository.GetChoices(ctx, dbConn, questionID)
		if err != nil {
			t.Fatal(err)
		}

		if len(choices) != 3 {
			t.Fatalf("expected 3 choices, got %d", len(choices))
		}

		// Verify ordering
		for i := 0; i < len(choices)-1; i++ {
			if choices[i].OrderNum >= choices[i+1].OrderNum {
				t.Error("choices should be ordered by order_num ASC")
			}
		}
	})

	// GIVEN: A choice exists
	// WHEN: Getting choice by ID
	// THEN: Returns correct choice
	t.Run("get choice by id", func(t *testing.T) {
		choices, _ := cRepository.GetChoices(ctx, dbConn, questionID)
		firstChoice := choices[0]

		fetched, err := cRepository.GetChoiceByID(ctx, dbConn, firstChoice.ID)
		if err != nil {
			t.Fatalf("failed to get choice: %v", err)
		}

		if fetched.Text != firstChoice.Text {
			t.Errorf("text: got %q, want %q", fetched.Text, firstChoice.Text)
		}
	})

	// GIVEN: A choice with duplicate order_num
	// WHEN: Creating choice
	// THEN: Should fail with UNIQUE constraint
	t.Run("duplicate order_num fails", func(t *testing.T) {
		duplicate := db.Choice{
			QuestionID: questionID,
			Text:       "Duplicate",
			OrderNum:   1, // Already exists
		}

		_, err := cRepository.CreateChoice(ctx, dbConn, duplicate)
		if err == nil {
			t.Error("expected error for duplicate order_num, got nil")
		}
	})

	// GIVEN: A question is deleted
	// WHEN: CASCADE delete triggers
	// THEN: Choices are deleted automatically
	t.Run("CASCADE delete question removes choices", func(t *testing.T) {
		// Create temp question with choices
		tempQuestion := db.Question{
			SessionID:     session.Id.String(),
			Text:          "Temp question",
			OrderNum:      2,
			AllowMultiple: false,
			MaxChoices:    1,
		}
		qRepository.CreateQuestion(ctx, dbConn, tempQuestion)

		tempQuestions, _ := qRepository.GetQuestions(ctx, dbConn, session.Id.String())
		var tempQuestionID int
		for _, q := range tempQuestions {
			if q.Text == "Temp question" {
				tempQuestionID = q.ID
				break
			}
		}

		tempChoice := db.Choice{
			QuestionID: tempQuestionID,
			Text:       "Temp choice",
			OrderNum:   1,
		}
		cRepository.CreateChoice(ctx, dbConn, tempChoice)

		// Delete question
		err := qRepository.DeleteQuestion(ctx, dbConn, tempQuestionID)
		if err != nil {
			t.Fatal(err)
		}

		// Verify choices were CASCADE deleted
		choices, err := cRepository.GetChoices(ctx, dbConn, tempQuestionID)
		if err != nil {
			t.Fatal(err)
		}

		if len(choices) != 0 {
			t.Errorf("expected 0 choices after CASCADE delete, got %d", len(choices))
		}
	})

	t.Run("delete choice", func(t *testing.T) {
		// Create temp question with choices
		tempQuestion := db.Question{
			SessionID:     session.Id.String(),
			Text:          "Temp question",
			OrderNum:      2,
			AllowMultiple: false,
			MaxChoices:    2,
		}

		questionID, err := qRepository.CreateQuestion(ctx, dbConn, tempQuestion)
		if err != nil {
			t.Fatal(err)
		}

		tempChoices := []db.Choice{
			{
				QuestionID: questionID,
				Text:       "Temp choice 1",
				OrderNum:   1,
			},
			{
				QuestionID: questionID,
				Text:       "Temp choice 2",
				OrderNum:   2,
			},
		}
		var ids []int
		for _, choice := range tempChoices {
			if id, err := cRepository.CreateChoice(ctx, dbConn, choice); err != nil {
				t.Fatal(err)
			} else {
				ids = append(ids, id)
			}
		}

		for _, id := range ids {

			if err := cRepository.DeleteChoice(ctx, dbConn, id); err != nil {
				t.Fatal(err)
			}
		}

	})
}
