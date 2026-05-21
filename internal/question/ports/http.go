package ports

import (
	"net/http"

	"github.com/73NN0/voting-app/internal/common/server"
	"github.com/73NN0/voting-app/internal/common/server/httperr"
	"github.com/73NN0/voting-app/internal/question/app"
)

const (
	invalidJSON string = "invalid JSON"
)

// TODO : getChoiceByID
// TODO TEST  later : test nobody should be able to get questions & choices from another session
type HttpHandler struct {
	service *app.Service
}

func NewHttpHandler(service *app.Service) *HttpHandler {
	return &HttpHandler{
		service: service,
	}
}

// funny
func (h *HttpHandler) teapot(w http.ResponseWriter, r *http.Request) {
	httperr.Teapot(w, "teapot, would you like some tea sir ?")
}

func AddRoutes(r *server.Router, h *HttpHandler) {
	r.Group("/questions", func(sub *server.Router) {

		// URL: GET /questions/ or GET /questions/anything (catch-all)
		sub.Handle("GET /{$}", server.Chain(
			http.HandlerFunc(h.teapot),
			server.Logging, server.Recovery, server.CORS,
		))

		// CRUD Questions
		// URL: POST /questions
		sub.Handle("POST /", server.Chain(
			http.HandlerFunc(h.CreateQuestion),
			server.Logging, server.Recovery, server.CORS,
		))

		// URL: GET /questions/{id}
		sub.Handle("GET /{id}", server.Chain(
			http.HandlerFunc(h.GetQuestionByID),
			server.Logging, server.Recovery, server.CORS,
		))

		// URL: PUT /questions/{id}
		sub.Handle("PUT /{id}", server.Chain(
			http.HandlerFunc(h.UpdateQuestion),
			server.Logging, server.Recovery, server.CORS,
		))

		// URL: DELETE /questions/{id}
		sub.Handle("DELETE /{id}", server.Chain(
			http.HandlerFunc(h.DeleteQuestion),
			server.Logging, server.Recovery, server.CORS,
		))

		// Liste par session
		// URL: GET /questions/session/{sessionID}
		sub.Handle("GET /session/{sessionID}", server.Chain(
			http.HandlerFunc(h.ListQuestionsBySessionID),
			server.Logging, server.Recovery, server.CORS,
		))

		// CRUD Choices (nested under a question)
		// URL: POST /questions/{questionID}/choices
		sub.Handle("POST /{questionID}/choices", server.Chain(
			http.HandlerFunc(h.CreateChoice),
			server.Logging, server.Recovery, server.CORS,
		))

		// URL: GET /questions/{questionID}/choices
		sub.Handle("GET /{questionID}/choices", server.Chain(
			http.HandlerFunc(h.ListChoicesByQuestionID),
			server.Logging, server.Recovery, server.CORS,
		))

		// URL: PUT /questions/{questionID}/choices/{choiceID}
		sub.Handle("PUT /{questionID}/choices/{choiceID}", server.Chain(
			http.HandlerFunc(h.UpdateChoice),
			server.Logging, server.Recovery, server.CORS,
		))

		// URL: DELETE /questions/{questionID}/choices/{choiceID}
		sub.Handle("DELETE /{questionID}/choices/{choiceID}", server.Chain(
			http.HandlerFunc(h.DeleteChoice),
			server.Logging, server.Recovery, server.CORS,
		))
	})
}
