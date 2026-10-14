package web

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) UnmarkAttendanceAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	coach, ok := h.requireCoach(w, r)
	if !ok {
		return
	}

	var req struct {
		TrainingID int `json:"training_id"`
		StudentID  int `json:"student_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if req.TrainingID <= 0 || req.StudentID <= 0 {
		http.Error(w, "training_id and student_id required", http.StatusBadRequest)
		return
	}

	training, err := h.scheduleService.GetTrainingByID(req.TrainingID)
	if err != nil || training == nil {
		http.Error(w, "Training not found", http.StatusNotFound)
		return
	}
	coachProfile, err := h.coachService.GetCoachByUserID(coach.ID)
	if err != nil || coachProfile == nil {
		http.Error(w, "Coach profile not found", http.StatusBadRequest)
		return
	}
	if training.CoachID == nil || *training.CoachID != coachProfile.ID {
		http.Error(w, "Not your training", http.StatusForbidden)
		return
	}

	if err := h.attendanceService.UnmarkAttendance(req.TrainingID, req.StudentID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
}
