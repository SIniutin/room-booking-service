package http

import (
	"encoding/json"
	"net/http"

	"github.com/SIniutin/room-booking-service/internal/dto"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return
	}
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, dto.ErrorResponse{
		Error: dto.ErrorDTO{
			Code:    code,
			Message: message,
		},
	})
}
