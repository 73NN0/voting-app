package db_test

import (
	"context"
	"database/sql/driver"
	"testing"
	"time"

	db "github.com/73NN0/voting-app/db"
	_ "modernc.org/sqlite"
)

func TestTimestamp_Scan(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    time.Time
		wantErr bool
	}{
		{
			name:  "valid RFC3339 string",
			input: "2024-01-15T19:05:00Z",
			want:  time.Date(2024, 1, 15, 19, 5, 0, 0, time.UTC),
		},
		{
			name:  "nil value",
			input: nil,
			want:  time.Time{},
		},
		{
			name:    "invalid type",
			input:   123,
			wantErr: true,
		},
		{
			name:    "invalid date string",
			input:   "not-a-date",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ts db.Timestamp
			err := ts.Scan(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !ts.Time.Equal(tt.want) {
				t.Errorf("got %v, want %v", ts.Time, tt.want)
			}
		})
	}
}

func TestTimestamp_Value(t *testing.T) {
	tests := []struct {
		name    string
		ts      db.Timestamp
		want    driver.Value
		wantErr bool
	}{
		{
			name: "valid db.Timestamp",
			ts:   db.Timestamp{time.Date(2024, 1, 15, 19, 5, 0, 0, time.UTC)},
			want: "2024-01-15T19:05:00Z",
		},
		{
			name: "zero time",
			ts:   db.Timestamp{time.Time{}},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.ts.Value()

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTimestamp_Integration(t *testing.T) {
	database := db.NewSQLiteDBRepository()
	defer database.OpenDB(":memory:")()

	if err := database.InitializeDatabaseSchemas(); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	sdb := database.GetDB()

	// Test INSERT (Value() appelé)
	t.Run("insert db.Timestamp", func(t *testing.T) {
		now := db.Timestamp{time.Date(2024, 1, 15, 19, 5, 0, 0, time.UTC)}

		_, err := sdb.ExecContext(ctx, `
            CREATE TABLE test_Timestamps (id INTEGER PRIMARY KEY, created_at TEXT)
        `)
		if err != nil {
			t.Fatal(err)
		}

		_, err = sdb.ExecContext(ctx, `
            INSERT INTO test_Timestamps (id, created_at) VALUES (1, ?)
        `, now)
		if err != nil {
			t.Fatalf("failed to insert: %v", err)
		}

		// Vérifier le stockage
		var stored string
		err = sdb.QueryRowContext(ctx, `
            SELECT created_at FROM test_Timestamps WHERE id = 1
        `).Scan(&stored)

		if err != nil {
			t.Fatal(err)
		}

		want := "2024-01-15T19:05:00Z"
		if stored != want {
			t.Errorf("stored as %q, want %q", stored, want)
		}
	})

	// Test SELECT (Scan() appelé)
	t.Run("scan db.Timestamp", func(t *testing.T) {
		var ts db.Timestamp

		err := sdb.QueryRowContext(ctx, `
            SELECT created_at FROM test_Timestamps WHERE id = 1
        `).Scan(&ts)
		if err != nil {
			t.Fatalf("failed to scan: %v", err)
		}

		want := time.Date(2024, 1, 15, 19, 5, 0, 0, time.UTC)
		if !ts.Time.Equal(want) {
			t.Errorf("scanned %v, want %v", ts.Time, want)
		}
	})

	// Test round-trip
	t.Run("round trip", func(t *testing.T) {
		original := db.Timestamp{time.Now().UTC().Truncate(time.Second)}

		_, err := sdb.ExecContext(ctx, `
            INSERT INTO test_Timestamps (id, created_at) VALUES (2, ?)
        `, original)
		if err != nil {
			t.Fatal(err)
		}

		var retrieved db.Timestamp
		err = sdb.QueryRowContext(ctx, `
            SELECT created_at FROM test_Timestamps WHERE id = 2
        `).Scan(&retrieved)
		if err != nil {
			t.Fatal(err)
		}

		if !retrieved.Time.Equal(original.Time) {
			t.Errorf("got %v, want %v", retrieved.Time, original.Time)
		}
	})
}

// TODO : fix it after Session

// func TestSession_WithTimestamp(t *testing.T) {
// 	database := db.NewSQLiteDBRepository()
// 	defer database.OpenDB("sqlite", ":memory:")()

// 	if err := database.InitializeDatabaseSchemas(); err != nil {
// 		t.Fatal(err)
// 	}

// 	ctx := context.Background()

// 	sdb := database.GetDB()

// 	sRepository := db.NewVoteSessionRepository(db)

// 	session := Session{
// 		Id:          uuid.New(),
// 		Title:       "Test Session",
// 		Description: "Test",
// 		CreatedAt:   time.Now().UTC(),
// 		EndsAt:      time.Date(2026, 2, 15, 18, 0, 0, 0, time.UTC),
// 	}

// 	// Insert
// 	err := sRepository.CreateVoteSession(ctx, session)
// 	if err != nil {
// 		t.Fatalf("failed to create: %v", err)
// 	}

// 	// Retrieve
// 	fetched, err := sRepository.GetVoteSessionByID(ctx, session.Id)
// 	if err != nil {
// 		t.Fatalf("failed to get: %v", err)
// 	}

// 	// Compare (les timestamps devraient être identiques)
// 	if !fetched.EndsAt.Equal(session.EndsAt) {
// 		t.Errorf("EndsAt: got %v, want %v", fetched.EndsAt, session.EndsAt)
// 	}
// }

// func TestTimestamp_EdgeCases(t *testing.T) {
// 	db, ctx, cleanup := setup(t, "sqlite", ":memory:")
// 	defer cleanup()

// 	sRepository := db.NewVoteSessionRepository(db)
// 	// GIVEN: A session with zero CreatedAt (relying on DB DEFAULT)
// 	// WHEN: Creating and fetching the session
// 	// THEN: CreatedAt should be populated by SQLite's datetime('now')
// 	//       Format: "2025-11-30 14:42:19" (SQLite default)
// 	//       and be parsable by Scan()
// 	t.Run("zero CreatedAt uses DB DEFAULT", func(t *testing.T) {
// 		session := entities.Session{
// 			Id:          uuid.New(),
// 			Title:       "Test",
// 			Description: "Test",
// 			CreatedAt:   time.Time{}, // Zero value - triggers DB DEFAULT
// 			EndsAt:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
// 		}

// 		err := sRepository.CreateVoteSession(ctx, session)
// 		if err != nil {
// 			t.Fatalf("failed to create session: %v", err)
// 		}

// 		fetched, err := sRepository.GetVoteSessionByID(ctx, session.Id)
// 		if err != nil {
// 			t.Fatalf("failed to fetch session: %v", err)
// 		}

// 		// SUCCESS: CreatedAt should be populated by DB
// 		if fetched.CreatedAt.IsZero() {
// 			t.Fatal("EXPECTED: CreatedAt populated by DB DEFAULT, GOT: zero value")
// 		}

// 		// SUCCESS: Should be recent (within 5 seconds)
// 		elapsed := time.Since(fetched.CreatedAt)
// 		if elapsed > 5*time.Second {
// 			t.Errorf("EXPECTED: CreatedAt recent (< 5s), GOT: %v ago", elapsed)
// 		}

// 		t.Logf("SUCCESS: DB DEFAULT created timestamp: %v", fetched.CreatedAt)
// 	})

// 	// GIVEN: A timestamp with exact seconds (no subseconds)
// 	//        Input: 2025-11-30T14:42:19Z (RFC3339)
// 	// WHEN: Round-tripping through Value() (RFC3339) and Scan()
// 	//       Stored as: "2025-11-30T14:42:19Z"
// 	// THEN: Timestamp should be preserved exactly
// 	//       Output: 2025-11-30T14:42:19Z (no precision loss)
// 	t.Run("exact seconds roundtrip - no precision loss", func(t *testing.T) {
// 		exactSecond := time.Date(2025, 11, 30, 14, 42, 19, 0, time.UTC)
// 		// Represents: "2025-11-30T14:42:19Z"

// 		session := entities.Session{
// 			Id:          uuid.New(),
// 			Title:       "Test",
// 			Description: "Test",
// 			CreatedAt:   exactSecond,
// 			EndsAt:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
// 		}

// 		err := sRepository.CreateVoteSession(ctx, session)
// 		if err != nil {
// 			t.Fatalf("failed to create session: %v", err)
// 		}

// 		fetched, err := sRepository.GetVoteSessionByID(ctx, session.Id)
// 		if err != nil {
// 			t.Fatalf("failed to fetch session: %v", err)
// 		}

// 		// SUCCESS: Exact match expected
// 		if !fetched.CreatedAt.Equal(session.CreatedAt) {
// 			t.Errorf("EXPECTED: exact match, GOT: %v, WANT: %v",
// 				fetched.CreatedAt, session.CreatedAt)
// 		} else {
// 			t.Logf("SUCCESS: Exact second preserved: %v", fetched.CreatedAt)
// 		}
// 	})

// 	// GIVEN: Raw SQL insert using SQLite's DEFAULT datetime('now')
// 	//        No explicit timestamp provided
// 	// WHEN: Fetching via Scan() which expects multiple formats
// 	//       DB stores as: "2025-11-30 14:42:19" (SQLite format - note space, no T, no Z)
// 	// THEN: SQLite format "2006-01-02 15:04:05" should be parsed correctly
// 	//       Scan() should handle: "2025-11-30 14:42:19" → time.Time
// 	t.Run("parse SQLite DEFAULT format", func(t *testing.T) {
// 		sessionID := uuid.New()

// 		// Direct SQL insert - uses SQLite's datetime('now') format
// 		_, err := db.ExecContext(ctx, `
//             INSERT INTO vote_session (id, title, description)
//             VALUES (?, 'Test', 'Test')
//         `, sessionID)
// 		if err != nil {
// 			t.Fatalf("failed to insert: %v", err)
// 		}

// 		// Fetch and parse
// 		fetched, err := sRepository.GetVoteSessionByID(ctx, sessionID)
// 		if err != nil {
// 			t.Fatalf("failed to parse SQLite format: %v", err)
// 		}

// 		// SUCCESS: Should parse SQLite's "YYYY-MM-DD HH:MM:SS" format
// 		if fetched.CreatedAt.IsZero() {
// 			t.Fatal("EXPECTED: SQLite datetime parsed, GOT: zero value")
// 		}

// 		// Should be recent
// 		if time.Since(fetched.CreatedAt) > 5*time.Second {
// 			t.Error("EXPECTED: recent timestamp, GOT: stale timestamp")
// 		}

// 		t.Logf("SUCCESS: Parsed SQLite DEFAULT format: %v", fetched.CreatedAt)
// 	})

// 	// GIVEN: Two sessions - one with SQLite DEFAULT, one with Go RFC3339
// 	//        Session 1: DB generates "2025-11-30 14:42:19" (SQLite format)
// 	//        Session 2: Go provides "2025-11-30T14:42:19Z" (RFC3339 format)
// 	// WHEN: Storing both formats in DB
// 	//       Format difference: space vs 'T', no 'Z' vs 'Z' suffix
// 	// THEN: Both should be parsable, formats differ but semantic meaning preserved
// 	//       Both → valid time.Time via Scan()
// 	t.Run("compare storage formats in DB", func(t *testing.T) {
// 		// Session 1: SQLite DEFAULT format
// 		session1ID := uuid.New()
// 		_, err := db.ExecContext(ctx, `
//             INSERT INTO vote_session (id, title, description)
//             VALUES (?, 'SQLite Format', 'Test')
//         `, session1ID)
// 		if err != nil {
// 			t.Fatalf("failed to insert session1: %v", err)
// 		}

// 		// Session 2: Go RFC3339 format
// 		session2 := entities.Session{
// 			Id:          uuid.New(),
// 			Title:       "RFC3339 Format",
// 			Description: "Test",
// 			CreatedAt:   time.Now().UTC(),
// 			EndsAt:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
// 		}
// 		err = sRepository.CreateVoteSession(ctx, session2)
// 		if err != nil {
// 			t.Fatalf("failed to insert session2: %v", err)
// 		}

// 		// Inspect raw DB formats
// 		var format1, format2 string
// 		err = db.QueryRowContext(ctx,
// 			"SELECT created_at FROM vote_session WHERE id = ?",
// 			session1ID,
// 		).Scan(&format1)
// 		if err != nil {
// 			t.Fatalf("failed to read format1: %v", err)
// 		}

// 		err = db.QueryRowContext(ctx,
// 			"SELECT created_at FROM vote_session WHERE id = ?",
// 			session2.Id.String(),
// 		).Scan(&format2)
// 		if err != nil {
// 			t.Fatalf("failed to read format2: %v", err)
// 		}

// 		t.Logf("Format in DB:")
// 		t.Logf("  SQLite DEFAULT: %q (example: \"2025-11-30 14:42:19\")", format1)
// 		t.Logf("  Go RFC3339:     %q (example: \"2025-11-30T14:42:19Z\")", format2)

// 		// EXPECTED: Formats differ (this is OK)
// 		if format1 == format2 {
// 			t.Error("UNEXPECTED: Both formats identical (should differ)")
// 		}

// 		// SUCCESS: Both should be parsable by Scan()
// 		fetched1, err := sRepository.GetVoteSessionByID(ctx, session1ID)
// 		if err != nil {
// 			t.Fatalf("EXPECTED: SQLite format parsable, GOT error: %v", err)
// 		}

// 		fetched2, err := sRepository.GetVoteSessionByID(ctx, session2.Id)
// 		if err != nil {
// 			t.Fatalf("EXPECTED: RFC3339 format parsable, GOT error: %v", err)
// 		}

// 		// Both should be valid timestamps
// 		if fetched1.CreatedAt.IsZero() {
// 			t.Error("EXPECTED: valid timestamp from SQLite format, GOT: zero")
// 		}
// 		if fetched2.CreatedAt.IsZero() {
// 			t.Error("EXPECTED: valid timestamp from RFC3339 format, GOT: zero")
// 		}

// 		t.Logf("SUCCESS: Both formats parsed correctly")
// 		t.Logf("  SQLite format → %v", fetched1.CreatedAt)
// 		t.Logf("  RFC3339 format → %v", fetched2.CreatedAt)

// 		// EXPECTED: Timestamps should be close (within 2 seconds of each other)
// 		diff := fetched1.CreatedAt.Sub(fetched2.CreatedAt)
// 		if diff < 0 {
// 			diff = -diff
// 		}
// 		if diff > 2*time.Second {
// 			t.Errorf("EXPECTED: timestamps within 2s, GOT: %v apart", diff)
// 		}
// 	})

// 	// GIVEN: Using RFC3339Nano format (preserves nanoseconds)
// 	//        Input: 2025-11-30T14:42:19.123456789Z (full precision)
// 	// WHEN: Round-tripping with nanosecond precision
// 	//       If Value() uses RFC3339:     "2025-11-30T14:42:19Z" (nanos lost)
// 	//       If Value() uses RFC3339Nano: "2025-11-30T14:42:19.123456789Z" (nanos kept)
// 	// THEN: If Value() uses RFC3339Nano, nanos preserved; otherwise lost
// 	//       Test documents which behavior is active
// 	t.Run("RFC3339 vs RFC3339Nano - document precision", func(t *testing.T) {
// 		withNanos := time.Date(2025, 11, 30, 14, 42, 19, 123456789, time.UTC)
// 		// Represents: "2025-11-30T14:42:19.123456789Z" (9 digits)

// 		session := entities.Session{
// 			Id:          uuid.New(),
// 			Title:       "Nano Test",
// 			Description: "Test",
// 			CreatedAt:   withNanos,
// 			EndsAt:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
// 		}

// 		err := sRepository.CreateVoteSession(ctx, session)
// 		if err != nil {
// 			t.Fatalf("failed to create: %v", err)
// 		}

// 		// Check raw format in DB
// 		var rawFormat string
// 		err = db.QueryRowContext(ctx,
// 			"SELECT created_at FROM vote_session WHERE id = ?",
// 			session.Id.String(),
// 		).Scan(&rawFormat)
// 		if err != nil {
// 			t.Fatalf("failed to read format: %v", err)
// 		}

// 		t.Logf("Stored format: %q", rawFormat)
// 		t.Logf("  If RFC3339:     expect \"2025-11-30T14:42:19Z\" (no fractional seconds)")
// 		t.Logf("  If RFC3339Nano: expect \"2025-11-30T14:42:19.123456789Z\" (with .nnnnnnnnn)")

// 		fetched, err := sRepository.GetVoteSessionByID(ctx, session.Id)
// 		if err != nil {
// 			t.Fatalf("failed to fetch: %v", err)
// 		}

// 		originalNanos := session.CreatedAt.Nanosecond()
// 		fetchedNanos := fetched.CreatedAt.Nanosecond()

// 		if originalNanos == fetchedNanos {
// 			t.Logf("SUCCESS: Nanoseconds preserved (RFC3339Nano in use)")
// 			t.Logf("  Original: %d nanos", originalNanos)
// 			t.Logf("  Fetched:  %d nanos", fetchedNanos)
// 		} else {
// 			t.Logf("EXPECTED LOSS: Nanoseconds truncated (RFC3339 in use)")
// 			t.Logf("  Original: %d nanos", originalNanos)
// 			t.Logf("  Fetched:  %d nanos (lost)", fetchedNanos)

// 			// Verify seconds still match
// 			if !fetched.CreatedAt.Truncate(time.Second).Equal(session.CreatedAt.Truncate(time.Second)) {
// 				t.Errorf("EXPECTED: seconds match, GOT mismatch")
// 			}
// 		}
// 	})
// }
