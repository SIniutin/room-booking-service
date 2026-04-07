package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"go.uber.org/zap"

	d "github.com/internships-backend/test-backend-SIniutin/internal/domain"
	"github.com/internships-backend/test-backend-SIniutin/internal/dto"
)

func (h *Handler) dummyLogin(w http.ResponseWriter, r *http.Request) {
	var req dto.DummyLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Warn("dummyLogin: decode body", zap.Error(err))
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request")
		return
	}

	if req.Role == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "role is required")
		return
	}

	t, err := h.auth.DummyLogin(r.Context(), d.UserRole(req.Role))
	if err != nil {
		h.log.Warn("dummyLogin: usecase", zap.String("role", req.Role), zap.Error(err))
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	h.log.Info("dummyLogin: ok", zap.String("role", req.Role))
	writeJSON(w, http.StatusOK, dto.DummyLoginResponse{Token: string(t)})
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Warn("register: decode body", zap.Error(err))
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request")
		return
	}

	if req.Email == "" || req.Password == "" || req.Role == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "email, password and role are required")
		return
	}

	user, err := h.auth.Register(r.Context(), req.Email, req.Password, d.UserRole(req.Role))
	if err != nil {
		if errors.Is(err, d.ErrUserAlreadyExists) {
			writeError(w, http.StatusConflict, "USER_ALREADY_EXISTS", err.Error())
			return
		}
		h.log.Warn("register: usecase", zap.String("email", req.Email), zap.Error(err))
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	h.log.Info("register: ok", zap.String("user_id", uuid.UUID(user.ID).String()), zap.String("email", req.Email))
	writeJSON(w, http.StatusCreated, dto.RegisterResponse{
		User: dto.UserDTO{
			ID:    uuid.UUID(user.ID).String(),
			Email: string(user.Email),
			Role:  string(user.Role),
		},
	})
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Warn("login: decode body", zap.Error(err))
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request")
		return
	}

	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "email and password are required")
		return
	}

	t, err := h.auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		h.log.Warn("login: usecase", zap.String("email", req.Email), zap.Error(err))
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
		return
	}

	h.log.Info("login: ok", zap.String("email", req.Email))
	writeJSON(w, http.StatusOK, dto.LoginResponse{Token: string(t)})
}
