package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/73NN0/voting-app/internal/common/db"
	"github.com/73NN0/voting-app/internal/common/server"
	"github.com/73NN0/voting-app/internal/questions/adapters"
	"github.com/73NN0/voting-app/internal/questions/app"
	"github.com/73NN0/voting-app/internal/questions/ports"
	sessions "github.com/73NN0/voting-app/internal/sessions/adapters"
)

// Note : later implement the switch logic if their are different type of ports
func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	dsn := flag.String("dsn", "voting.db", "sqlite data source name")
	flag.Parse()

	// TODO : entry point or volume docker ?
	database, cleanup, err := db.OpenSQLite(*dsn)
	if err != nil {
		log.Fatal(err)
	}

	defer cleanup()

	if err = db.InitializeSchemas(database); err != nil {
		log.Fatal(err)
	}

	questionsRepo := adapters.NewSqliteQuestionsRepository(database)
	choicesRepo := adapters.NewSqliteChoicesRepositoy(database)

	sessionsRepo := sessions.NewSqliteSessionRepository(database)

	sessionsChecker := adapters.NewSessionCheckerInProcess(sessionsRepo)

	service := app.NewService(questionsRepo, choicesRepo, sessionsChecker)

	router := server.NewRouter()

	ports.AddRoutes(router, ports.NewHttpHandler(service))

	http.ListenAndServe(*addr, router.Handler())
}
