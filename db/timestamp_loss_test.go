//go:build precision_tests
// +build precision_tests

// create a separate package for timestamp
package db_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	vdb "gitlab.com/singfield/voting-app/db"
)

func TestTimestamp_EdgeCases_loss_precision_time(t *testing.T) {
	db, ctx, cleanup := setup(t, "sqlite", ":memory:")
	defer cleanup()
	// GIVEN: A timestamp with nanosecond precision
	//        Input: 2025-11-30T14:42:19.123456789Z (Go time.Time)
	// WHEN: Using RFC3339 format (which only preserves seconds)
	//       Value() formats as: "2025-11-30T14:42:19Z" (nanos lost)
	// THEN: Nanoseconds WILL BE LOST (this is expected behavior)
	//       Output: 2025-11-30T14:42:19Z (123456789 nanos → 0 nanos)
	t.Run("nanoseconds roundtrip - EXPECTED precision loss", func(t *testing.T) {
		withNanos := time.Date(2025, 11, 30, 14, 42, 19, 123456789, time.UTC)
		// Represents: "2025-11-30T14:42:19.123456789Z"

		session := vdb.Session{
			Id:          uuid.New(),
			Title:       "Test",
			Description: "Test",
			CreatedAt:   vdb.Timestamp{withNanos},
			EndsAt:      vdb.Timestamp{time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		}

		err := vdb.CreateVoteSession(ctx, db, session)
		if err != nil {
			t.Fatalf("failed to create session: %v", err)
		}

		fetched, err := vdb.GetVoteSessionByID(ctx, db, session.Id.String())
		if err != nil {
			t.Fatalf("failed to fetch session: %v", err)
		}

		// EXPECTED LOSS: Nanoseconds should be truncated
		if fetched.CreatedAt.Equal(session.CreatedAt.Time) {
			t.Error("UNEXPECTED: nanoseconds preserved (should be lost with RFC3339)")
		}

		// Verify seconds match (nano loss is acceptable)
		if !fetched.CreatedAt.Truncate(time.Second).Equal(session.CreatedAt.Truncate(time.Second)) {
			t.Fatalf("EXPECTED: seconds match, GOT: %v, WANT: %v (at second precision)",
				fetched.CreatedAt, session.CreatedAt)
		}

		t.Logf("SUCCESS: Nanosecond precision lost as expected")
		t.Logf("  Original: %v (nanos: %d)", session.CreatedAt, session.CreatedAt.Nanosecond())
		t.Logf("  Retrieved: %v (nanos: %d)", fetched.CreatedAt, fetched.CreatedAt.Nanosecond())
	})
}
