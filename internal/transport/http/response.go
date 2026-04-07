package http

import (
	"encoding/json"
	"net/http"

	"github.com/internships-backend/test-backend-SIniutin/internal/dto"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, dto.ErrorResponse{
		Error: dto.ErrorDTO{
			Code:    code,
			Message: message,
		},
	})
}
