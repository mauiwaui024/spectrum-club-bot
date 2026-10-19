package web

import (
	"encoding/json"
	"errors"
	"net/http"

	"spectrum-club-bot/internal/service"
)

func (h *Handler) RegisterStudentByCoachAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	coach, ok := h.requireCoach(w, r)
	if !ok {
		return
	}
	var req struct {
		TrainingID int   `json:"training_id"`
		StudentID  int64 `json:"student_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if req.TrainingID <= 0 || req.StudentID <= 0 {
		http.Error(w, "Укажите тренировку и ученика", http.StatusBadRequest)
		return
	}
	if err := h.attendanceService.SignUpStudentByCoach(coach.ID, req.StudentID, req.TrainingID); err != nil {
		http.Error(w, err.Error(), coachSignupStatus(err))
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func coachSignupStatus(err error) int {
	switch {
	case errors.Is(err, service.ErrNotTrainingCoach):
		return http.StatusForbidden
	case errors.Is(err, service.ErrTrainingNotFound), errors.Is(err, service.ErrStudentNotFound):
		return http.StatusNotFound
	case errors.Is(err, service.ErrNotAStudent),
		errors.Is(err, service.ErrInactiveStudent),
		errors.Is(err, service.ErrNoSubscription),
		errors.Is(err, service.ErrAlreadyRegistered):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
