package ports

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/73NN0/voting-app/internal/common/logger"
	"github.com/73NN0/voting-app/internal/common/server/httperr"
	"github.com/73NN0/voting-app/internal/common/server/httpstat"
	"github.com/73NN0/voting-app/internal/question/app"
	"github.com/73NN0/voting-app/internal/question/domain/question"
	"github.com/google/uuid"
)

const (
	createQuestionFailed string = "create question failed"
	validateFailed       string = "validation failed"
	listQuestionFailed   string = "list questions failed"
	updateQuestionFailed string = "update question failed"
	deleteQuestionFailed string = "delete question failed"
)

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

func (h *HttpHandler) CreateQuestion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req questionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Logger.Error(invalidJSON, "err", err)
		httperr.BadRequest(w, invalidJSON)
		return
	}

	if err := Validate(req); err != nil {
		logger.Logger.Warn(validateFailed, "err", err)
		httperr.BadRequest(w, err.Error())
		return
	}

	id, err := h.service.CreateQuestion(ctx, req.SessionID, req.Text, req.OrderNum, req.MaxChoices, req.AllowMultiple)
	if err != nil {
		if errors.Is(err, app.ErrVoteSessionNotFound) {
			httperr.NotFound(w, app.ErrVoteSessionNotFound.Error())
			return
		}
		logger.Logger.Error(createQuestionFailed, "err", err)
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
		logger.Logger.Warn(question.InvalidQuestionID, "id", idStr)
		httperr.BadRequest(w, question.InvalidQuestionID)
		return
	}

	q, err := h.service.GetQuestionByID(ctx, id)
	if err != nil {
		if errors.Is(err, question.ErrNotFound) {
			logger.Logger.Warn(question.NotFound, "id", id)
			httperr.NotFound(w, question.NotFound)
			return
		}
		logger.Logger.Error("get question failed", "err", err)
		httperr.InternalServerError(w, httperr.InternalServerErrorMsg)
		return
	}

	httpstat.OkJSON(w, q)
}

func (h *HttpHandler) ListQuestionsBySessionID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := r.PathValue("sessionID")
	sessionID, err := uuid.Parse(idStr)
	if err != nil {
		logger.Logger.Warn(question.InvalidSessionID, "id", idStr)
		httperr.BadRequest(w, question.InvalidSessionID)
		return
	}

	questions, err := h.service.ListQuestionsBySessionID(ctx, sessionID)
	if err != nil {
		logger.Logger.Error(listQuestionFailed, "err", err, "session_id", sessionID)
		httperr.InternalServerError(w, httperr.InternalServerErrorMsg)
		return
	}
	if questions == nil {
		questions = []question.Question{} // [] instead of "null" in JSON
	}

	httpstat.OkJSON(w, questions)
}

func (h *HttpHandler) UpdateQuestion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Logger.Warn(question.InvalidQuestionID, "id", idStr)
		httperr.BadRequest(w, question.InvalidQuestionID)
		return
	}

	var req questionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Logger.Error(invalidJSON, "err", err)
		httperr.BadRequest(w, invalidJSON)
		return
	}

	q, err := h.service.GetQuestionByID(ctx, id)
	if err != nil {
		if errors.Is(err, question.ErrNotFound) {
			logger.Logger.Warn(question.NotFound)
			httperr.NotFound(w, question.NotFound)
			return
		}

		logger.Logger.Error(updateQuestionFailed, "id", id, "err", err)
		httperr.InternalServerError(w, httperr.InternalServerErrorMsg)
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

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Logger.Warn(question.InvalidQuestionID, "id", idStr)
		httperr.BadRequest(w, question.InvalidQuestionID)
		return
	}

	if err := h.service.DeleteQuestion(ctx, id); err != nil {
		if errors.Is(err, question.ErrNotFound) {
			logger.Logger.Warn(question.NotFound, "id", id)
			httperr.NotFound(w, question.NotFound)
			return
		}

		logger.Logger.Error(deleteQuestionFailed, "err", err)
		httperr.InternalServerError(w, httperr.InternalServerErrorMsg)
		return
	}

	httpstat.NoContent(w)
}
