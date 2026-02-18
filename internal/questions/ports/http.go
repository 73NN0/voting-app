package ports

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/73NN0/voting-app/internal/common/logger"
	"github.com/73NN0/voting-app/internal/common/server"
	"github.com/73NN0/voting-app/internal/common/server/httperr"
	"github.com/73NN0/voting-app/internal/common/server/httpstat"
	"github.com/73NN0/voting-app/internal/questions/app"
	"github.com/google/uuid"
)

type HttpHandler struct {
	service *app.Service
}

type questionRequest struct {
	SessionID     uuid.UUID `json:"session_id"`
	Text          string    `json:"text"`
	OrderNum      int       `json:"order_num"`
	MaxChoices    int       `json:"max_choices"`
	AllowMultiple bool      `json:"allow_multiple"`
}

func Validate(req questionRequest) error {
	if req.SessionID == uuid.Nil {
		return errors.New("session_id is required")
	}

	if req.Text == "" {
		return errors.New("text is required")
	}

	if req.OrderNum <= 0 {
		return errors.New("order_num must be positive")
	}

	if req.MaxChoices < 1 && req.AllowMultiple {
		return errors.New("max_choices must be positive when allow_multiple is true")
	}

	return nil
}

func NewHttpHandler(service *app.Service) *HttpHandler {
	return &HttpHandler{
		service: service,
	}
}

func (h *HttpHandler) CreateQuestion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req questionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Logger.Error("invalid JSON", "err", err)
		httperr.BadRequest(w, "invalid JSON")
		return
	}

	if err := Validate(req); err != nil {
		logger.Logger.Warn("validation failed", "err", err)
		httperr.BadRequest(w, err.Error())
		return
	}

	id, err := h.service.CreateQuestion(ctx, req.SessionID, req.Text, req.OrderNum, req.MaxChoices, req.AllowMultiple)
	if err != nil {
		logger.Logger.Error("create question failed", "err", err)
		httperr.InternalServerError(w, err.Error())
		return
	}

	httpstat.CreatedJSON(w, strconv.Itoa(id))
}

func (h *HttpHandler) GetQuestionByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)

	if err != nil {
		logger.Logger.Warn("invalid question ID", "idStr", idStr)
		httperr.BadRequest(w, "invalid question ID")
		return
	}

	q, err := h.service.GetQuestionByID(ctx, id)
	if err != nil {
		logger.Logger.Error("get question failed", "err", err)
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
		logger.Logger.Warn("invalid session ID", "idStr", idStr)
		httperr.BadRequest(w, "invalid ID")
		return
	}

	questions, err := h.service.ListQuestionsBySessionID(ctx, sessionID)
	if err != nil {
		logger.Logger.Error("list questions failed", "err", err)
		httperr.NotFound(w, "questions for this session not found")
	}

	httpstat.OkJSON(w, questions)
}

func (h *HttpHandler) UpdateQuestion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Logger.Warn("invalid question ID", "idStr", idStr)
		httperr.BadRequest(w, "invalid question ID")
		return
	}

	var req questionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Logger.Error("invalid JSON", "err", err)
		httperr.BadRequest(w, "invalid JSON")
		return
	}

	q, err := h.service.GetQuestionByID(ctx, id)
	if err != nil {
		logger.Logger.Error("get question failed", "err", err)
		httperr.NotFound(w, "question not found")
		return
	}

	q.UpdateText(req.Text)
	q.ChangeOrderNum(req.OrderNum)
	q.UpdateMaxChoices(req.MaxChoices)
	q.UpdateAllowMultiple(req.AllowMultiple)

	if err := h.service.UpdateQuestion(ctx, q.ID(), q.Text(), q.OrderNum(), q.MaxChoices(), q.AllowMultiple()); err != nil {
		logger.Logger.Error("update question failed", "err", err)
		httperr.InternalServerError(w, err.Error())
		return
	}

	httpstat.OkJSON(w, q)
}

func (h *HttpHandler) DeleteQuestion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := r.URL.Path[1:] // ou r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Logger.Warn("invalid question ID", "idStr", idStr)
		httperr.BadRequest(w, "invalid question ID")
		return
	}

	if err := h.service.DeleteQuestion(ctx, id); err != nil {
		logger.Logger.Error("delete question failed", "err", err)
		httperr.NotFound(w, "question not found")
		return
	}

	httpstat.NoContent(w, "question deleted")
}

func (h *HttpHandler) CreateChoice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := r.PathValue("questionID")
	questionID, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Logger.Warn("invalid question ID", "idStr", idStr)
		httperr.BadRequest(w, "invalid question ID")
		return
	}

	var req struct {
		Text     string `json:"text"`
		OrderNum int    `json:"order_num"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Logger.Error("invalid JSON", "err", err)
		httperr.BadRequest(w, "invalid JSON")
		return
	}

	id, err := h.service.CreateChoice(ctx, questionID, req.OrderNum, req.Text)
	if err != nil {
		logger.Logger.Error("create choice failed", "err", err)
		httperr.BadRequest(w, err.Error())
		return
	}

	httpstat.CreatedJSON(w, map[string]int{"id": id})
}

// funny
func (h *HttpHandler) teapot(w http.ResponseWriter, r *http.Request) {
	httperr.Teapot(w, "teapot, would you like some tea sir ?")
}

func (h *HttpHandler) ListChoicesByQuestionID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := r.URL.Path[1:] // ou r.PathValue("questionID") si Go 1.22
	questionID, err := strconv.Atoi(idStr)
	if err != nil {
		httperr.BadRequest(w, "invalid question ID")
		return
	}

	choices, err := h.service.ListChoicesByQuestionID(ctx, questionID)
	if err != nil {
		httperr.NotFound(w, "choices not found")
		return
	}

	httpstat.OkJSON(w, choices)
}

// UpdateChoice (exemple avec body JSON pour text/orderNum)
type updateChoiceRequest struct {
	Text     string `json:"text"`
	OrderNum int    `json:"order_num"`
}

func ValidateUpdateChoice(req updateChoiceRequest) error {
	if req.Text == "" {
		return errors.New("text is required")
	}
	if req.OrderNum <= 0 {
		return errors.New("order_num must be positive")
	}
	return nil
}

func (h *HttpHandler) UpdateChoice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := r.URL.Path[1:] // ou r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		httperr.BadRequest(w, "invalid choice ID")
		return
	}

	var req updateChoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.BadRequest(w, "invalid JSON")
		return
	}

	if err := ValidateUpdateChoice(req); err != nil {
		httperr.BadRequest(w, err.Error())
		return
	}
	// FIXME
	c, err := h.service.UpdateChoice(ctx, id, req.OrderNum, req.Text)
	if err != nil {
		httperr.NotFound(w, "choice not found")
		return
	}

	httpstat.OkJSON(w, c)
}

// DeleteChoice
func (h *HttpHandler) DeleteChoice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := r.URL.Path[1:] // ou r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		httperr.BadRequest(w, "invalid choice ID")
		return
	}

	if err := h.service.DeleteChoice(ctx, id); err != nil {
		httperr.NotFound(w, "choice not found")
		return
	}

	httpstat.NoContent(w, "choice deleted")
}

func AddRoutes(r *server.Router, h *HttpHandler) {
	r.Group("/questions", func(sub *server.Router) {
		sub.Handle("/{$}", server.Chain(
			http.HandlerFunc(h.teapot),
			server.Logging, server.Recovery, server.CORS,
		))

		sub.Handle("GET /session/{sessionID}", server.Chain(
			http.HandlerFunc(h.ListQuestionsBySessionID),
			server.Logging, server.Recovery, server.CORS,
		))

		sub.Handle("/{id}", server.Chain(
			http.HandlerFunc(h.GetQuestionByID),
			server.Logging, server.Recovery, server.CORS,
		))

	})
}
