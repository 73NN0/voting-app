package httpstat

import (
	"encoding/json"
	"net/http"
)

type StatusResponse struct {
	Status string `json:"status"`
}

func stat(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func OkJSON(w http.ResponseWriter, v interface{}) {
	stat(w, http.StatusOK, v)
}

// 200 ok
func Ok(w http.ResponseWriter, msg string) {
	stat(w, http.StatusOK, StatusResponse{Status: msg})
}

func CreatedJSON(w http.ResponseWriter, v interface{}) {
	stat(w, http.StatusCreated, v)
}

// 201
func Created(w http.ResponseWriter, msg string) {
	stat(w, http.StatusCreated, StatusResponse{Status: msg})
}

func AcceptedJSON(w http.ResponseWriter, v interface{}) {
	stat(w, http.StatusAccepted, v)
}

// 202
func Accepted(w http.ResponseWriter, msg string) {
	stat(w, http.StatusAccepted, StatusResponse{Status: msg})
}

// 204 - "A 204 response MUST NOT include any message body". RFC 9110
// [RFC](https://datatracker.ietf.org/doc/html/rfc9110#name-204-no-content)
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
