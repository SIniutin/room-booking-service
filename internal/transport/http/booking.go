package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	d "github.com/internships-backend/test-backend-SIniutin/internal/domain"
	"github.com/internships-backend/test-backend-SIniutin/internal/dto"
	m "github.com/internships-backend/test-backend-SIniutin/internal/transport/http/middleware"
)

func (h *Handler) createBooking(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(m.ClaimsKey).(m.Claims)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	var req dto.CreateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Warn("createBooking: decode body", zap.Error(err))
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request")
		return
	}

	slotID, err := uuid.Parse(req.SlotID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid slotId")
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid token")
		return
	}

	booking, err := h.booking.CreateBooking(r.Context(), d.UserID(userID), d.SlotID(slotID), req.CreateConferenceLink)
	if err != nil {
		if errors.Is(err, d.ErrSlotNotFound) {
			writeError(w, http.StatusNotFound, "SLOT_NOT_FOUND", err.Error())
			return
		}
		if errors.Is(err, d.ErrSlotAlreadyBooked) {
			writeError(w, http.StatusConflict, "SLOT_ALREADY_BOOKED", err.Error())
			return
		}
		if errors.Is(err, d.ErrSlotInThePast) {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
			return
		}
		h.log.Error("createBooking: usecase", zap.String("user_id", claims.UserID), zap.String("slot_id", req.SlotID), zap.Error(err))
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	h.log.Info("createBooking: ok", zap.String("booking_id", uuid.UUID(booking.ID).String()), zap.String("user_id", claims.UserID))
	writeJSON(w, http.StatusCreated, dto.CreateBookingResponse{Booking: bookingToDTO(booking)})
}

func (h *Handler) listBookings(w http.ResponseWriter, r *http.Request) {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if err != nil || pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	bookings, total, err := h.booking.ListBookings(r.Context(), page, pageSize)
	if err != nil {
		h.log.Error("listBookings: usecase", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	result := make([]dto.BookingDTO, len(bookings))
	for i, b := range bookings {
		result[i] = bookingToDTO(b)
	}

	writeJSON(w, http.StatusOK, dto.ListBookingsResponse{
		Bookings: result,
		Pagination: dto.PaginationDTO{
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	})
}

func (h *Handler) listMyBookings(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(m.ClaimsKey).(m.Claims)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid token")
		return
	}

	bookings, err := h.booking.ListMyBookings(r.Context(), d.UserID(userID))
	if err != nil {
		h.log.Error("listMyBookings: usecase", zap.String("user_id", claims.UserID), zap.Error(err))
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	result := make([]dto.BookingDTO, len(bookings))
	for i, b := range bookings {
		result[i] = bookingToDTO(b)
	}

	writeJSON(w, http.StatusOK, dto.ListMyBookingsResponse{Bookings: result})
}

func (h *Handler) cancelBooking(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(m.ClaimsKey).(m.Claims)
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	bookingID, err := uuid.Parse(chi.URLParam(r, "bookingId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid bookingId")
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid token")
		return
	}

	booking, err := h.booking.CancelBooking(r.Context(), d.UserID(userID), d.BookingID(bookingID))
	if err != nil {
		if errors.Is(err, d.ErrBookingNotFound) {
			writeError(w, http.StatusNotFound, "BOOKING_NOT_FOUND", err.Error())
			return
		}
		if errors.Is(err, d.ErrBookingForbidden) {
			writeError(w, http.StatusForbidden, "FORBIDDEN", err.Error())
			return
		}
		h.log.Error("cancelBooking: usecase", zap.String("booking_id", bookingID.String()), zap.String("user_id", claims.UserID), zap.Error(err))
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	h.log.Info("cancelBooking: ok", zap.String("booking_id", bookingID.String()), zap.String("user_id", claims.UserID))
	writeJSON(w, http.StatusOK, dto.CancelBookingResponse{Booking: bookingToDTO(booking)})
}

func bookingToDTO(b d.Booking) dto.BookingDTO {
	result := dto.BookingDTO{
		ID:     uuid.UUID(b.ID).String(),
		SlotID: uuid.UUID(b.SlotID).String(),
		UserID: uuid.UUID(b.UserID).String(),
		Status: string(b.Status),
	}
	if b.Link != "" {
		result.ConferenceLink = &b.Link
	}
	if !b.CreatedAt.IsZero() {
		result.CreatedAt = &b.CreatedAt
	}
	return result
}
