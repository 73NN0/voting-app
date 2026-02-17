package httperr

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func Error(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{Error: msg})
}

func BadRequest(w http.ResponseWriter, msg string) {
	Error(w, msg, http.StatusBadRequest)
}

func NotFound(w http.ResponseWriter, msg string) {
	Error(w, msg, http.StatusNotFound)
}

func InternalServerError(w http.ResponseWriter, msg string) {
	Error(w, msg, http.StatusInternalServerError)
}

// note : why not use this against bot ?
// in Honey pot ?
func Teapot(w http.ResponseWriter, msg string) {
	Error(w, msg, http.StatusTeapot) // because it's funny
}
