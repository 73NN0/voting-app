package questions_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/73NN0/voting-app/db"
	"github.com/73NN0/voting-app/internal/questions"
	choice "github.com/73NN0/voting-app/internal/questions/domain/choice"
	question "github.com/73NN0/voting-app/internal/questions/domain/question"
	"github.com/73NN0/voting-app/internal/sessions"
	session "github.com/73NN0/voting-app/internal/sessions/domain/session"
)

func TestQuestionRepository_CreateAndGet(t *testing.T) {
	database := db.NewSQLiteDBRepository()
	defer database.OpenDB(":memory:")()

	if err := database.InitializeDatabaseSchemas(); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	sessionRepo := sessions.NewSqliteSessionRepository(database)
	s, _ := session.NewSessionNoEnd("Test Session", "Test")

	err := sessionRepo.CreateVoteSession(ctx, s)
	if err != nil {
		t.Fatal(err)
	}

	repo := questions.NewSqliteQuestionsRepository(database)

	// GIVEN: Une nouvelle question du domaine
	q, err := question.NewQuestion(
		s.ID(),
		"Approuvez-vous le budget?",
		1,     // orderNum
		1,     // maxChoices,
		false, // allowMultiple
	)
	if err != nil {
		t.Fatal(err)
	}

	// WHEN: On la sauvegarde
	id, err := repo.CreateQuestion(ctx, &q)
	if err != nil {
		t.Fatalf("CreateQuestion failed: %v", err)
	}

	if id <= 0 {
		t.Errorf("expected id > 0, got %d", id)
	}

	// THEN: On peut la récupérer
	fetched, err := repo.GetQuestionByID(ctx, id)
	if err != nil {
		t.Fatalf("GetQuestionByID failed: %v", err)
	}

	// Vérifications
	if fetched.Text() != "Approuvez-vous le budget?" {
		t.Errorf("text: got %q, want %q", fetched.Text(), "Approuvez-vous le budget?")
	}

	if fetched.SessionID() != s.ID() {
		t.Errorf("sessionID: got %s, want %s", fetched.SessionID(), s.ID())
	}

	if fetched.OrderNum() != 1 {
		t.Errorf("orderNum: got %d, want 1", fetched.OrderNum())
	}

	if fetched.AllowMultiple() != false {
		t.Error("allowMultiple: got true, want false")
	}

	if fetched.MaxChoices() != 1 {
		t.Errorf("maxChoices: got %d, want 1", fetched.MaxChoices())
	}

	// CreatedAt devrait être récent
	if fetched.CreatedAt().IsZero() {
		t.Error("createdAt should not be zero")
	}
}

func TestQuestionRepository_GetQuestionsBySessionID(t *testing.T) {
	database := db.NewSQLiteDBRepository()
	defer database.OpenDB(":memory:")()

	if err := database.InitializeDatabaseSchemas(); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	repo := questions.NewSqliteQuestionsRepository(database)

	// GIVEN: Plusieurs questions dans une session
	sessionID := uuid.New()

	q1, _ := question.NewQuestion(sessionID, "Question 1", 1, 1, false)
	q2, _ := question.NewQuestion(sessionID, "Question 2", 2, 3, true)
	q3, _ := question.NewQuestion(sessionID, "Question 3", 3, 1, false)

	repo.CreateQuestion(ctx, &q1)
	repo.CreateQuestion(ctx, &q2)
	repo.CreateQuestion(ctx, &q3)

	// WHEN: On récupère toutes les questions de la session
	questions, err := repo.GetQuestionsBySessionID(ctx, sessionID)
	if err != nil {
		t.Fatalf("GetQuestionsBySessionID failed: %v", err)
	}

	// THEN: On a 3 questions ordonnées
	if len(questions) != 3 {
		t.Fatalf("expected 3 questions, got %d", len(questions))
	}

	// Vérifier l'ordre
	if questions[0].OrderNum() != 1 {
		t.Error("questions should be ordered by order_num")
	}
	if questions[1].OrderNum() != 2 {
		t.Error("questions should be ordered by order_num")
	}
	if questions[2].OrderNum() != 3 {
		t.Error("questions should be ordered by order_num")
	}

	// Vérifier les textes
	if questions[0].Text() != "Question 1" {
		t.Errorf("q1 text: got %q, want 'Question 1'", questions[0].Text())
	}
	if questions[1].Text() != "Question 2" {
		t.Errorf("q2 text: got %q, want 'Question 2'", questions[1].Text())
	}
}

func TestChoiceRepository_CreateAndGet(t *testing.T) {
	database := db.NewSQLiteDBRepository()
	defer database.OpenDB(":memory:")()

	if err := database.InitializeDatabaseSchemas(); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	repo := questions.NewSqliteQuestionsRepository(database)

	sessionRepo := sessions.NewSqliteSessionRepository(database)
	s, _ := session.NewSessionNoEnd("Test Session", "Test")

	err := sessionRepo.CreateVoteSession(ctx, s)
	if err != nil {
		t.Fatal(err)
	}
	q, _ := question.NewQuestion(s.ID(), "Élire le président", 1, 1, false)
	questionID, _ := repo.CreateQuestion(ctx, &q)

	choice1 := choice.NewChoice(questionID, 1, "Candidat A")
	choice2 := choice.NewChoice(questionID, 2, "Candidat B")
	choice3 := choice.NewChoice(questionID, 3, "Candidat C")

	// WHEN: On crée les choices
	id1, err := repo.CreateChoice(ctx, &choice1)
	if err != nil {
		t.Fatalf("CreateChoice failed: %v", err)
	}

	repo.CreateChoice(ctx, &choice2)
	repo.CreateChoice(ctx, &choice3)

	// THEN: On peut les récupérer individuellement
	fetched, err := repo.GetChoiceByID(ctx, id1)
	if err != nil {
		t.Fatalf("GetChoiceByID failed: %v", err)
	}

	if fetched.Text() != "Candidat A" {
		t.Errorf("text: got %q, want 'Candidat A'", fetched.Text())
	}

	// THEN: On peut récupérer tous les choices d'une question
	choices, err := repo.GetChoicesByQuestionID(ctx, questionID)
	if err != nil {
		t.Fatalf("GetChoicesByQuestionID failed: %v", err)
	}

	if len(choices) != 3 {
		t.Fatalf("expected 3 choices, got %d", len(choices))
	}

	// Vérifier l'ordre
	for i := 0; i < len(choices)-1; i++ {
		if choices[i].OrderNum() >= choices[i+1].OrderNum() {
			t.Error("choices should be ordered by order_num ASC")
		}
	}
}

func TestQuestionRepository_Delete(t *testing.T) {
	database := db.NewSQLiteDBRepository()
	defer database.OpenDB(":memory:")()

	if err := database.InitializeDatabaseSchemas(); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	repo := questions.NewSqliteQuestionsRepository(database)

	// GIVEN: Une question
	sessionRepo := sessions.NewSqliteSessionRepository(database)
	s, _ := session.NewSessionNoEnd("Test Session", "Test")

	err := sessionRepo.CreateVoteSession(ctx, s)
	if err != nil {
		t.Fatal(err)
	}

	q, _ := question.NewQuestion(s.ID(), "À supprimer", 1, 1, false)
	id, _ := repo.CreateQuestion(ctx, &q)

	// WHEN: On la supprime
	err = repo.DeleteQuestion(ctx, id)
	if err != nil {
		t.Fatalf("DeleteQuestion failed: %v", err)
	}

	// THEN: Elle n'existe plus
	_, err = repo.GetQuestionByID(ctx, id)
	if err == nil {
		t.Error("expected error when fetching deleted question, got nil")
	}
}

func TestQuestionRepository_CascadeDelete(t *testing.T) {
	database := db.NewSQLiteDBRepository()
	defer database.OpenDB(":memory:")()

	if err := database.InitializeDatabaseSchemas(); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	repo := questions.NewSqliteQuestionsRepository(database)

	// GIVEN: Une question avec des choices
	sessionRepo := sessions.NewSqliteSessionRepository(database)
	s, _ := session.NewSessionNoEnd("Test Session", "Test")

	err := sessionRepo.CreateVoteSession(ctx, s)
	if err != nil {
		t.Fatal(err)
	}
	q, _ := question.NewQuestion(s.ID(), "Question", 1, 1, false)
	questionID, _ := repo.CreateQuestion(ctx, &q)

	choice1 := choice.NewChoice(questionID, 1, "Choice 1")
	choice2 := choice.NewChoice(questionID, 2, "Choice 2")

	repo.CreateChoice(ctx, &choice1)
	repo.CreateChoice(ctx, &choice2)

	// WHEN: On supprime la question
	err = repo.DeleteQuestion(ctx, questionID)
	if err != nil {
		t.Fatal(err)
	}

	// THEN: Les choices sont CASCADE supprimés
	choices, err := repo.GetChoicesByQuestionID(ctx, questionID)
	if err != nil {
		t.Fatal(err)
	}

	if len(choices) != 0 {
		t.Errorf("expected 0 choices after CASCADE delete, got %d", len(choices))
	}
}
