package http

import "net/http"

func (h *Handler) info(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
