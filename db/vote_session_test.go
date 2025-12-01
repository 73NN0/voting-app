package db_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"gitlab.com/singfield/voting-app/db"
)

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
