package middleware

import (
	"encoding/json"
	"net/http"
)

type errorDTO struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type errorResponse struct {
	Error errorDTO `json:"error"`
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errorResponse{
		Error: errorDTO{Code: code, Message: message},
	})
}
