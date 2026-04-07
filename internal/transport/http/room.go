package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	d "github.com/internships-backend/test-backend-SIniutin/internal/domain"
	"github.com/internships-backend/test-backend-SIniutin/internal/dto"
)

func (h *Handler) createRoom(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Warn("createRoom: decode body", zap.Error(err))
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "room name is required")
		return
	}

	room, err := h.room.CreateRoom(r.Context(), req.Name, req.Description, req.Capacity)
	if err != nil {
		h.log.Error("createRoom: usecase", zap.String("name", req.Name), zap.Error(err))
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	h.log.Info("createRoom: ok", zap.String("room_id", uuid.UUID(room.ID).String()), zap.String("name", room.Name))
	writeJSON(w, http.StatusCreated, dto.CreateRoomResponse{
		Room: roomToDTO(room),
	})
}

func (h *Handler) listRooms(w http.ResponseWriter, r *http.Request) {
	rooms, err := h.room.ListRooms(r.Context())
	if err != nil {
		h.log.Error("listRooms: usecase", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	result := make([]dto.RoomDTO, len(rooms))
	for i, room := range rooms {
		result[i] = roomToDTO(room)
	}

	writeJSON(w, http.StatusOK, dto.ListRoomsResponse{Rooms: result})
}

func (h *Handler) createSchedule(w http.ResponseWriter, r *http.Request) {
	roomID, err := uuid.Parse(chi.URLParam(r, "roomId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid roomId")
		return
	}

	var req dto.CreateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Warn("createSchedule: decode body", zap.String("room_id", roomID.String()), zap.Error(err))
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request")
		return
	}

	if slices.ContainsFunc(req.DaysOfWeek, func(day int) bool {
		return day < 1 || day > 7
	}) {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid day of week")
		return
	}

	schedule, err := h.room.CreateSchedule(r.Context(), d.RoomID(roomID), req.DaysOfWeek, req.StartTime, req.EndTime)
	if err != nil {
		if errors.Is(err, d.ErrScheduleAlreadyExists) {
			writeError(w, http.StatusConflict, "SCHEDULE_EXISTS", err.Error())
			return
		}
		if errors.Is(err, d.ErrRoomNotFound) {
			writeError(w, http.StatusNotFound, "ROOM_NOT_FOUND", err.Error())
			return
		}
		if errors.Is(err, d.ErrInvalidWorkday) ||
			errors.Is(err, d.ErrEmptyWorkTime) ||
			errors.Is(err, d.ErrInvalidWorkTimeFormat) ||
			errors.Is(err, d.ErrInvalidScheduleTime) {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
			return
		}
		h.log.Error("createSchedule: usecase", zap.String("room_id", roomID.String()), zap.Error(err))
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	h.log.Info("createSchedule: ok", zap.String("room_id", roomID.String()))
	writeJSON(w, http.StatusCreated, dto.CreateScheduleResponse{
		Schedule: scheduleToDTO(schedule),
	})
}

func (h *Handler) listSlots(w http.ResponseWriter, r *http.Request) {
	roomID, err := uuid.Parse(chi.URLParam(r, "roomId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid roomId")
		return
	}

	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "date is required")
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid date format, expected YYYY-MM-DD")
		return
	}

	slots, err := h.room.ListSlots(r.Context(), d.RoomID(roomID), date)
	if err != nil {
		if errors.Is(err, d.ErrRoomNotFound) {
			writeError(w, http.StatusNotFound, "ROOM_NOT_FOUND", err.Error())
			return
		}
		h.log.Error("listSlots: usecase", zap.String("room_id", roomID.String()), zap.String("date", dateStr), zap.Error(err))
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	result := make([]dto.SlotDTO, len(slots))
	for i, slot := range slots {
		result[i] = dto.SlotDTO{
			ID:     uuid.UUID(slot.ID).String(),
			RoomID: uuid.UUID(slot.RoomID).String(),
			Start:  slot.Start,
			End:    slot.End,
		}
	}

	writeJSON(w, http.StatusOK, dto.ListSlotsResponse{Slots: result})
}

func roomToDTO(room d.Room) dto.RoomDTO {
	r := dto.RoomDTO{
		ID:          uuid.UUID(room.ID).String(),
		Name:        room.Name,
		Description: room.Description,
		Capacity:    room.Capacity,
	}
	if !room.CreatedAt.IsZero() {
		r.CreatedAt = &room.CreatedAt
	}
	return r
}

func scheduleToDTO(s d.Schedule) dto.ScheduleDTO {
	return dto.ScheduleDTO{
		ID:         uuid.UUID(s.ID).String(),
		RoomID:     uuid.UUID(s.RoomID).String(),
		DaysOfWeek: s.Days.ToSlice(),
		StartTime:  time.Time(s.Start).Format("15:04"),
		EndTime:    time.Time(s.End).Format("15:04"),
	}
}
