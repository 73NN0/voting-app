package ports

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/73NN0/voting-app/internal/common/logger"
	"github.com/73NN0/voting-app/internal/common/server/httperr"
	"github.com/73NN0/voting-app/internal/common/server/httpstat"
	"github.com/73NN0/voting-app/internal/question/domain/choice"
	"github.com/73NN0/voting-app/internal/question/domain/question"
)

const (
	createChoiceFailed = "create choice failed"
	updateChoiceFailed = "update choice failed"
	deleteChoiceFailed = "delete choice failed"
)

// TODO : Choice Request ObJ !
func (h *HttpHandler) CreateChoice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := r.PathValue("questionID")
	questionID, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Logger.Warn(choice.InvalidQuestionID, "id", idStr)
		httperr.BadRequest(w, choice.InvalidQuestionID)
		return
	}

	var req struct {
		Text     string `json:"text"`
		OrderNum int    `json:"order_num"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Logger.Error(invalidJSON, "err", err)
		httperr.BadRequest(w, invalidJSON)
		return
	}

	id, err := h.service.CreateChoice(ctx, questionID, req.OrderNum, req.Text)
	if err != nil {
		if errors.Is(err, question.ErrInvalidQuestionID) {
			logger.Logger.Warn(question.InvalidQuestionID, "id", id, "question_id", questionID, "err", err, "text", req.Text)
			httperr.BadRequest(w, question.InvalidQuestionID)
			return
		}
		logger.Logger.Error(createChoiceFailed, "err", err)
		httperr.InternalServerError(w, httperr.InternalServerErrorMsg)
		return
	}

	httpstat.CreatedJSON(w, map[string]int{"id": id})
}

func (h *HttpHandler) ListChoicesByQuestionID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := r.PathValue("questionID")
	questionID, err := strconv.Atoi(idStr)
	if err != nil {
		httperr.BadRequest(w, "invalid question ID")
		return
	}

	choices, err := h.service.ListChoicesByQuestionID(ctx, questionID)
	if err != nil {
		logger.Logger.Error("list choices failed", "err", err, "question_id", questionID)
		httperr.InternalServerError(w, httperr.InternalServerErrorMsg)
		return
	}

	if choices == nil {
		choices = []choice.Choice{}
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

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		httperr.BadRequest(w, choice.InvalidChoiceID)
		return
	}

	var req updateChoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httperr.BadRequest(w, invalidJSON)
		return
	}

	if err := ValidateUpdateChoice(req); err != nil {
		httperr.BadRequest(w, err.Error())
		return
	}

	err = h.service.UpdateChoice(ctx, id, req.Text, req.OrderNum)
	if err != nil {
		if errors.Is(err, choice.ErrNotFound) {
			logger.Logger.Warn(choice.NotFound, "id", id)
			httperr.NotFound(w, choice.NotFound)
			return
		}
		logger.Logger.Error(updateChoiceFailed, "id", id, "err", err)
		httperr.InternalServerError(w, httperr.InternalServerErrorMsg)
		return
	}

	httpstat.NoContent(w)
}

// DeleteChoice
func (h *HttpHandler) DeleteChoice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		httperr.BadRequest(w, choice.InvalidQuestionID)
		return
	}

	if err := h.service.DeleteChoice(ctx, id); err != nil {
		if errors.Is(err, choice.ErrNotFound) {
			logger.Logger.Warn(choice.NotFound, "id", id)
			httperr.NotFound(w, choice.NotFound)
			return
		}
		logger.Logger.Error(deleteChoiceFailed, "id", id, "err", err)
		httperr.InternalServerError(w, httperr.InternalServerErrorMsg)
		return

	}

	httpstat.NoContent(w)
}
