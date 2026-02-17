package ports

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/73NN0/voting-app/internal/common/server"
	"github.com/73NN0/voting-app/internal/common/server/httperr"
	"github.com/73NN0/voting-app/internal/common/server/httpstat"
	"github.com/73NN0/voting-app/internal/questions/app"
	"github.com/google/uuid"
)

type HttpHandler struct {
	service *app.Service
}

func NewHttpHandler(service *app.Service) *HttpHandler {
	return &HttpHandler{
		service: service,
	}
}

func (h *HttpHandler) CreateQuestion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		SessionID     uuid.UUID `json:"session_id"`
		Text          string    `json:"text"`
		OrderNum      int       `json:"order_num"`
		MaxChoices    int       `json:"max_choices"`
		AllowMultiple bool      `json:"allow_multiple"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.BadRequest(w, "invalid JSON")
		return
	}

	id, err := h.service.CreateQuestion(ctx, req.SessionID, req.Text, req.OrderNum, req.MaxChoices, req.AllowMultiple)
	if err != nil {
		httperr.BadRequest(w, err.Error())
		return
	}

	httpstat.CreatedJSON(w, strconv.Itoa(id))
}

func (h *HttpHandler) GetQuestionByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// idea from letsgo
	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)

	if err != nil {
		httperr.BadRequest(w, "invalid question ID")
		return
	}

	q, err := h.service.GetQuestionByID(ctx, id)
	if err != nil {
		httperr.NotFound(w, "question not found")
		return
	}

	httpstat.OkJSON(w, q)
}

func (h *HttpHandler) ListQuestionsBySessionID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := r.PathValue("sessionID")
	sessionID, err := uuid.Parse(idStr)
	if err != nil {
		httperr.BadRequest(w, "invalid ID")
		return
	}

	questions, err := h.service.ListQuestionsBySessionID(ctx, sessionID)
	if err != nil {
		httperr.NotFound(w, "questions for this session not found")
	}

	httpstat.OkJSON(w, questions)
}

// funny
func (h *HttpHandler) teapot(w http.ResponseWriter, r *http.Request) {
	httperr.Teapot(w, "teapot, would you like some tea sir ?")
}

// TODO : Implement the rest of the crud

func AddRoutes(r *server.Router, h *HttpHandler) {
	r.Group("/questions", func(sub *server.Router) {
		sub.Handle("/{$}", server.Chain(
			http.HandlerFunc(h.teapot),
			server.Logging, server.Recovery, server.CORS,
		))

		sub.Handle("/session/{sessionID}", server.Chain(
			http.HandlerFunc(h.ListQuestionsBySessionID),
			server.Logging, server.Recovery, server.CORS,
		))

		sub.Handle("/{id}", server.Chain(
			http.HandlerFunc(h.GetQuestionByID),
			server.Logging, server.Recovery, server.CORS,
		))
	})
}
