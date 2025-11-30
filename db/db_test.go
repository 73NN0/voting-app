package db_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"gitlab.com/singfield/voting-app/db"

	_ "modernc.org/sqlite"
)

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

		assertStructEqual(t, user, *userQueried)
	})

	t.Run("get user by email", func(t *testing.T) {
		userQueried, err := db.GetUserByEmail(ctx, dbConn, user.Email)
		if err != nil {
			t.Fatal(err)
		}

		assertStructEqual(t, user, *userQueried)
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

		assertStructEqual(t, want, *userQueried)

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

	session := db.Session{
		Id:          uuid.New(),
		Title:       "Ravalement 2030",
		Description: "describing",
		CreatedAt:   db.Timestamp{time.Now().UTC()},
		EndsAt:      db.Timestamp{time.Date(2026, time.February, 15, 18, 0, 0, 0, time.UTC)},
	}

	t.Run("create vote session", func(t *testing.T) {
		if err := db.CreateVoteSession(ctx, dbConn, session); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("get vote session by id", func(t *testing.T) {
		fetched, err := db.GetVoteSessionByID(ctx, dbConn, session.Id.String())
		if err != nil {
			t.Fatal(err)
		}
		assertStructEqual(t, session, *fetched)
	})

	t.Run("update vote session", func(t *testing.T) {
		updated := session
		updated.Title = "Ravalement 2031"
		updated.Description = "new description"
		updated.EndsAt = db.Timestamp{time.Date(2027, time.March, 1, 12, 0, 0, 0, time.UTC)}

		err := db.UpdateVoteSession(ctx, dbConn, updated)
		if err != nil {
			t.Fatalf("failed to update session: %v", err)
		}

		fetched, err := db.GetVoteSessionByID(ctx, dbConn, session.Id.String())
		if err != nil {
			t.Fatal(err)
		}

		if fetched.Title != updated.Title {
			t.Errorf("title: got %q, want %q", fetched.Title, updated.Title)
		}
		if fetched.Description != updated.Description {
			t.Errorf("description: got %q, want %q", fetched.Description, updated.Description)
		}
	})

	t.Run("close vote session", func(t *testing.T) {
		beforeClose := time.Now().UTC()

		err := db.CloseVoteSession(ctx, dbConn, session.Id.String())
		if err != nil {
			t.Fatalf("failed to close session: %v", err)
		}

		fetched, err := db.GetVoteSessionByID(ctx, dbConn, session.Id.String())
		if err != nil {
			t.Fatal(err)
		}

		// EndsAt should be set to now
		if fetched.EndsAt.IsZero() {
			t.Error("EndsAt should be set")
		}

		// Should be recent (within 2 seconds)
		if fetched.EndsAt.Before(beforeClose) || time.Since(fetched.EndsAt.Time) > 2*time.Second {
			t.Errorf("EndsAt should be recent, got %v", fetched.EndsAt)
		}
	})

	t.Run("list vote sessions", func(t *testing.T) {
		// Create additional sessions
		session2 := db.Session{
			Id:          uuid.New(),
			Title:       "AG 2025",
			Description: "Annual meeting",
			CreatedAt:   db.Timestamp{time.Now().UTC()},
			EndsAt:      db.Timestamp{time.Date(2025, time.December, 31, 18, 0, 0, 0, time.UTC)},
		}
		session3 := db.Session{
			Id:          uuid.New(),
			Title:       "Budget 2026",
			Description: "Budget vote",
			CreatedAt:   db.Timestamp{time.Now().UTC()},
			EndsAt:      db.Timestamp{time.Date(2026, time.January, 15, 18, 0, 0, 0, time.UTC)},
		}

		db.CreateVoteSession(ctx, dbConn, session2)
		db.CreateVoteSession(ctx, dbConn, session3)

		// List first page (2 items)
		sessions, err := db.ListVoteSessions(ctx, dbConn, 2, 0)
		if err != nil {
			t.Fatalf("failed to list sessions: %v", err)
		}

		if len(sessions) != 2 {
			t.Errorf("expected 2 sessions, got %d", len(sessions))
		}

		// List second page (1 item)
		sessions, err = db.ListVoteSessions(ctx, dbConn, 2, 2)
		if err != nil {
			t.Fatalf("failed to list sessions (page 2): %v", err)
		}

		if len(sessions) != 1 {
			t.Errorf("expected 1 session on page 2, got %d", len(sessions))
		}

		// Sessions should be ordered by created_at DESC
		allSessions, _ := db.ListVoteSessions(ctx, dbConn, 10, 0)
		for i := 0; i < len(allSessions)-1; i++ {
			if allSessions[i].CreatedAt.Before(allSessions[i+1].CreatedAt.Time) {
				t.Error("sessions should be ordered by created_at DESC")
			}
		}
	})

	t.Run("delete vote session", func(t *testing.T) {
		// Create a session to delete
		toDelete := db.Session{
			Id:          uuid.New(),
			Title:       "To Delete",
			Description: "Will be deleted",
			CreatedAt:   db.Timestamp{time.Now().UTC()},
			EndsAt:      db.Timestamp{time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC)},
		}

		err := db.CreateVoteSession(ctx, dbConn, toDelete)
		if err != nil {
			t.Fatal(err)
		}

		// Delete it
		err = db.DeleteVoteSession(ctx, dbConn, toDelete.Id.String())
		if err != nil {
			t.Fatalf("failed to delete session: %v", err)
		}

		// Verify deletion
		_, err = db.GetVoteSessionByID(ctx, dbConn, toDelete.Id.String())
		if err == nil {
			t.Error("session should be deleted")
		}
	})
}

func TestSessionParticipants(t *testing.T) {
	dbConn, ctx, cleanup := setup(t, "sqlite", ":memory:")
	defer cleanup()

	// Create test users
	user1 := db.User{
		Id:    uuid.New(),
		Name:  "Alice",
		Email: "alice@test.com",
		Mdp:   "password1",
	}
	user2 := db.User{
		Id:    uuid.New(),
		Name:  "Bob",
		Email: "bob@test.com",
		Mdp:   "password2",
	}

	db.CreateUser(ctx, dbConn, user1)
	db.CreateUser(ctx, dbConn, user2)

	// Create test session
	session := db.Session{
		Id:          uuid.New(),
		Title:       "Test Session",
		Description: "For participants",
		CreatedAt:   db.Timestamp{time.Now().UTC()},
		EndsAt:      db.Timestamp{time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC)},
	}
	db.CreateVoteSession(ctx, dbConn, session)

	t.Run("add participant", func(t *testing.T) {
		err := db.AddParticipant(ctx, dbConn, session.Id.String(), user1.Id.String())
		if err != nil {
			t.Fatalf("failed to add participant: %v", err)
		}

		// Verify
		isParticipant, err := db.IsParticipant(ctx, dbConn, session.Id.String(), user1.Id.String())
		if err != nil {
			t.Fatal(err)
		}

		if !isParticipant {
			t.Error("user1 should be a participant")
		}
	})

	t.Run("add multiple participants", func(t *testing.T) {
		err := db.AddParticipant(ctx, dbConn, session.Id.String(), user2.Id.String())
		if err != nil {
			t.Fatalf("failed to add participant: %v", err)
		}

		// Get all participants
		participants, err := db.GetParticipants(ctx, dbConn, session.Id.String())
		if err != nil {
			t.Fatalf("failed to get participants: %v", err)
		}

		if len(participants) != 2 {
			t.Errorf("expected 2 participants, got %d", len(participants))
		}
	})

	t.Run("get user sessions", func(t *testing.T) {
		// Create another session for user1
		session2 := db.Session{
			Id:          uuid.New(),
			Title:       "Another Session",
			Description: "User1 only",
			CreatedAt:   db.Timestamp{time.Now().UTC()},
			EndsAt:      db.Timestamp{time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC)},
		}
		db.CreateVoteSession(ctx, dbConn, session2)
		db.AddParticipant(ctx, dbConn, session2.Id.String(), user1.Id.String())

		// Get user1's sessions
		sessions, err := db.GetUserVoteSessions(ctx, dbConn, user1.Id.String())
		if err != nil {
			t.Fatalf("failed to get user sessions: %v", err)
		}

		if len(sessions) != 2 {
			t.Errorf("user1 should have 2 sessions, got %d", len(sessions))
		}

		// Get user2's sessions
		sessions, err = db.GetUserVoteSessions(ctx, dbConn, user2.Id.String())
		if err != nil {
			t.Fatal(err)
		}

		if len(sessions) != 1 {
			t.Errorf("user2 should have 1 session, got %d", len(sessions))
		}
	})

	t.Run("remove participant", func(t *testing.T) {
		err := db.RemoveParticipant(ctx, dbConn, session.Id.String(), user2.Id.String())
		if err != nil {
			t.Fatalf("failed to remove participant: %v", err)
		}

		// Verify removal
		isParticipant, err := db.IsParticipant(ctx, dbConn, session.Id.String(), user2.Id.String())
		if err != nil {
			t.Fatal(err)
		}

		if isParticipant {
			t.Error("user2 should not be a participant anymore")
		}

		// user1 should still be there
		participants, err := db.GetParticipants(ctx, dbConn, session.Id.String())
		if err != nil {
			t.Fatal(err)
		}

		if len(participants) != 1 {
			t.Errorf("expected 1 participant remaining, got %d", len(participants))
		}
	})

	t.Run("is not participant", func(t *testing.T) {
		user3 := db.User{
			Id:    uuid.New(),
			Name:  "Charlie",
			Email: "charlie@test.com",
			Mdp:   "password3",
		}
		db.CreateUser(ctx, dbConn, user3)

		isParticipant, err := db.IsParticipant(ctx, dbConn, session.Id.String(), user3.Id.String())
		if err != nil {
			t.Fatal(err)
		}

		if isParticipant {
			t.Error("user3 should not be a participant")
		}
	})

	t.Run("CASCADE delete session removes participants", func(t *testing.T) {
		// Create new session with participant
		tempSession := db.Session{
			Id:          uuid.New(),
			Title:       "Temp Session",
			Description: "Will be deleted",
			CreatedAt:   db.Timestamp{time.Now().UTC()},
			EndsAt:      db.Timestamp{time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)},
		}
		db.CreateVoteSession(ctx, dbConn, tempSession)
		db.AddParticipant(ctx, dbConn, tempSession.Id.String(), user1.Id.String())

		// Delete session
		err := db.DeleteVoteSession(ctx, dbConn, tempSession.Id.String())
		if err != nil {
			t.Fatal(err)
		}

		// Verify participant entry was CASCADE deleted
		isParticipant, err := db.IsParticipant(ctx, dbConn, tempSession.Id.String(), user1.Id.String())
		if err != nil {
			t.Fatal(err)
		}

		if isParticipant {
			t.Error("participant entry should be CASCADE deleted")
		}
	})
}

func TestQuestion(t *testing.T) {
	dbConn, ctx, cleanup := setup(t, "sqlite", ":memory:")
	defer cleanup()

	// Setup: Create a session first
	session := db.Session{
		Id:          uuid.New(),
		Title:       "AG 2025",
		Description: "Annual meeting",
		CreatedAt:   db.Timestamp{time.Now().UTC()},
		EndsAt:      db.Timestamp{time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC)},
	}
	err := db.CreateVoteSession(ctx, dbConn, session)
	if err != nil {
		t.Fatal(err)
	}

	// GIVEN: A session exists
	// WHEN: Creating a question
	// THEN: Question is stored with correct order_num
	t.Run("create question", func(t *testing.T) {
		question := db.Question{
			SessionID:     session.Id.String(),
			Text:          "Approuvez-vous le budget 2025?",
			OrderNum:      1,
			AllowMultiple: false,
			MaxChoices:    1,
		}

		err := db.CreateQuestion(ctx, dbConn, question)
		if err != nil {
			t.Fatalf("failed to create question: %v", err)
		}

		// Verify question was created
		questions, err := db.GetQuestions(ctx, dbConn, session.Id.String())
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
		question2 := db.Question{
			SessionID:     session.Id.String(),
			Text:          "Élisez le nouveau président",
			OrderNum:      2,
			AllowMultiple: false,
			MaxChoices:    1,
		}

		question3 := db.Question{
			SessionID:     session.Id.String(),
			Text:          "Choix multiples test",
			OrderNum:      3,
			AllowMultiple: true,
			MaxChoices:    3,
		}

		err := db.CreateQuestion(ctx, dbConn, question2)
		if err != nil {
			t.Fatal(err)
		}

		err = db.CreateQuestion(ctx, dbConn, question3)
		if err != nil {
			t.Fatal(err)
		}

		// Get all questions
		questions, err := db.GetQuestions(ctx, dbConn, session.Id.String())
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
		questions, _ := db.GetQuestions(ctx, dbConn, session.Id.String())
		firstQuestion := questions[0]

		fetched, err := db.GetQuestionByID(ctx, dbConn, firstQuestion.ID)
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
		duplicate := db.Question{
			SessionID:     session.Id.String(),
			Text:          "Duplicate order",
			OrderNum:      1, // Already exists
			AllowMultiple: false,
			MaxChoices:    1,
		}

		err := db.CreateQuestion(ctx, dbConn, duplicate)
		if err == nil {
			t.Error("expected error for duplicate order_num, got nil")
		}
	})

	// GIVEN: A session is deleted
	// WHEN: CASCADE delete triggers
	// THEN: Questions are deleted automatically
	t.Run("CASCADE delete session removes questions", func(t *testing.T) {
		// Create temp session with question
		tempSession := db.Session{
			Id:          uuid.New(),
			Title:       "Temp",
			Description: "Will be deleted",
			CreatedAt:   db.Timestamp{time.Now().UTC()},
			EndsAt:      db.Timestamp{time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC)},
		}
		db.CreateVoteSession(ctx, dbConn, tempSession)

		tempQuestion := db.Question{
			SessionID:     tempSession.Id.String(),
			Text:          "Temp question",
			OrderNum:      1,
			AllowMultiple: false,
			MaxChoices:    1,
		}
		db.CreateQuestion(ctx, dbConn, tempQuestion)

		// Delete session
		err := db.DeleteVoteSession(ctx, dbConn, tempSession.Id.String())
		if err != nil {
			t.Fatal(err)
		}

		// Verify questions were CASCADE deleted
		questions, err := db.GetQuestions(ctx, dbConn, tempSession.Id.String())
		if err != nil {
			t.Fatal(err)
		}

		if len(questions) != 0 {
			t.Errorf("expected 0 questions after CASCADE delete, got %d", len(questions))
		}
	})
}

func TestChoice(t *testing.T) {
	dbConn, ctx, cleanup := setup(t, "sqlite", ":memory:")
	defer cleanup()

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
	db.CreateQuestion(ctx, dbConn, question)

	// Get question ID
	questions, _ := db.GetQuestions(ctx, dbConn, session.Id.String())
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

		err := db.CreateChoice(ctx, dbConn, choice)
		if err != nil {
			t.Fatalf("failed to create choice: %v", err)
		}

		// Verify
		choices, err := db.GetChoices(ctx, dbConn, questionID)
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

		db.CreateChoice(ctx, dbConn, choice2)
		db.CreateChoice(ctx, dbConn, choice3)

		choices, err := db.GetChoices(ctx, dbConn, questionID)
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
		choices, _ := db.GetChoices(ctx, dbConn, questionID)
		firstChoice := choices[0]

		fetched, err := db.GetChoiceByID(ctx, dbConn, firstChoice.ID)
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

		err := db.CreateChoice(ctx, dbConn, duplicate)
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
		db.CreateQuestion(ctx, dbConn, tempQuestion)

		tempQuestions, _ := db.GetQuestions(ctx, dbConn, session.Id.String())
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
		db.CreateChoice(ctx, dbConn, tempChoice)

		// Delete question
		err := db.DeleteQuestion(ctx, dbConn, tempQuestionID)
		if err != nil {
			t.Fatal(err)
		}

		// Verify choices were CASCADE deleted
		choices, err := db.GetChoices(ctx, dbConn, tempQuestionID)
		if err != nil {
			t.Fatal(err)
		}

		if len(choices) != 0 {
			t.Errorf("expected 0 choices after CASCADE delete, got %d", len(choices))
		}
	})
}
func assertStructEqual(t *testing.T, want, got interface{}) {
	t.Helper()
	if !reflect.DeepEqual(want, got) {
		t.Errorf("got %+v,\n want \n %+v \n\r", got, want)
	}
}
