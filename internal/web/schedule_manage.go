package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"spectrum-club-bot/internal/models"
)

func (h *Handler) GroupsAPI(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireCoach(w, r); !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		groups, err := h.groupService.GetAllGroups()
		if err != nil {
			http.Error(w, "Failed to load groups", http.StatusInternalServerError)
			return
		}
		if groups == nil {
			groups = []models.TrainingGroup{}
		}
		h.writeJSON(w, http.StatusOK, map[string]interface{}{"groups": groups})
	case http.MethodPost:
		h.createGroupAPI(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) CoachScheduleAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, ok := h.requireCoach(w, r)
	if !ok {
		return
	}
	coach, err := h.coachService.GetCoachByUserID(user.ID)
	if err != nil || coach == nil {
		http.Error(w, "Coach profile not found", http.StatusBadRequest)
		return
	}

	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")
	start := time.Now()
	// Default covers max from-templates horizon (12 weeks) + current week.
	end := start.AddDate(0, 0, 12*7+7)
	if startStr != "" {
		if t, err := time.Parse("2006-01-02", startStr); err == nil {
			start = t
		}
	}
	if endStr != "" {
		if t, err := time.Parse("2006-01-02", endStr); err == nil {
			end = t
		}
	}

	items, err := h.scheduleService.GetCoachSchedule(coach.ID, start, end)
	if err != nil {
		http.Error(w, "Failed to load schedule", http.StatusInternalServerError)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]interface{}{"trainings": items})
}

func (h *Handler) TrainingsAPI(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createTrainingAPI(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) TrainingByIDAPI(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/api/trainings/")
	rest = strings.Trim(rest, "/")
	if rest == "from-templates" {
		h.TrainingsFromTemplatesAPI(w, r)
		return
	}
	if strings.HasSuffix(rest, "/plan") {
		idPart := strings.TrimSuffix(rest, "/plan")
		idPart = strings.Trim(idPart, "/")
		id, err := strconv.Atoi(idPart)
		if err != nil || strings.Contains(idPart, "/") {
			http.Error(w, "Invalid training id", http.StatusBadRequest)
			return
		}
		h.trainingPlanAPI(w, r, id)
		return
	}
	if rest == "" || strings.Contains(rest, "/") {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	id, err := strconv.Atoi(rest)
	if err != nil {
		http.Error(w, "Invalid training id", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodPatch:
		h.patchTrainingAPI(w, r, id)
	case http.MethodDelete:
		h.deleteTrainingAPI(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) createTrainingAPI(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requireCoach(w, r)
	if !ok {
		return
	}
	coach, err := h.coachService.GetCoachByUserID(user.ID)
	if err != nil || coach == nil {
		http.Error(w, "Coach profile not found", http.StatusBadRequest)
		return
	}

	var req struct {
		GroupID     int    `json:"group_id"`
		Date        string `json:"date"`       // YYYY-MM-DD
		StartTime   string `json:"start_time"` // HH:MM
		EndTime     string `json:"end_time"`   // HH:MM
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if req.GroupID <= 0 || req.Date == "" || req.StartTime == "" || req.EndTime == "" {
		http.Error(w, "group_id, date, start_time, end_time required", http.StatusBadRequest)
		return
	}

	start, err := parseLocalDateTime(req.Date, req.StartTime)
	if err != nil {
		http.Error(w, "Invalid date/start_time", http.StatusBadRequest)
		return
	}
	end, err := parseLocalDateTime(req.Date, req.EndTime)
	if err != nil {
		http.Error(w, "Invalid date/end_time", http.StatusBadRequest)
		return
	}
	if !end.After(start) {
		http.Error(w, "end_time must be after start_time", http.StatusBadRequest)
		return
	}

	training := &models.TrainingSchedule{
		GroupID:      req.GroupID,
		CoachID:      &coach.ID,
		TrainingDate: start,
		StartTime:    start,
		EndTime:      end,
		Description:  req.Description,
		CreatedBy:    &user.ID,
	}
	if err := h.scheduleService.CreateTraining(training); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "training": training})
}

func (h *Handler) patchTrainingAPI(w http.ResponseWriter, r *http.Request, id int) {
	user, ok := h.requireCoach(w, r)
	if !ok {
		return
	}
	coach, err := h.coachService.GetCoachByUserID(user.ID)
	if err != nil || coach == nil {
		http.Error(w, "Coach profile not found", http.StatusBadRequest)
		return
	}
	existing, err := h.scheduleService.GetTrainingByID(id)
	if err != nil || existing == nil {
		http.Error(w, "Training not found", http.StatusNotFound)
		return
	}
	if existing.CoachID == nil || *existing.CoachID != coach.ID {
		http.Error(w, "Not your training", http.StatusForbidden)
		return
	}

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	updates := map[string]interface{}{}
	if v, ok := req["description"].(string); ok {
		updates["description"] = v
	}
	if dateStr, ok := req["date"].(string); ok && dateStr != "" {
		if startStr, ok2 := req["start_time"].(string); ok2 && startStr != "" {
			if t, err := time.ParseInLocation("2006-01-02 15:04", dateStr+" "+startStr, time.Local); err == nil {
				updates["training_date"] = t
				updates["start_time"] = t
			}
		}
		if endStr, ok2 := req["end_time"].(string); ok2 && endStr != "" {
			if t, err := time.ParseInLocation("2006-01-02 15:04", dateStr+" "+endStr, time.Local); err == nil {
				updates["end_time"] = t
			}
		}
	} else {
		dateBase := existing.TrainingDate.Format("2006-01-02")
		if startStr, ok := req["start_time"].(string); ok && startStr != "" {
			if t, err := time.ParseInLocation("2006-01-02 15:04", dateBase+" "+startStr, time.Local); err == nil {
				updates["start_time"] = t
			}
		}
		if endStr, ok := req["end_time"].(string); ok && endStr != "" {
			if t, err := time.ParseInLocation("2006-01-02 15:04", dateBase+" "+endStr, time.Local); err == nil {
				updates["end_time"] = t
			}
		}
	}
	if len(updates) == 0 {
		http.Error(w, "No updates", http.StatusBadRequest)
		return
	}
	if err := h.scheduleService.UpdateTrainingPartial(id, updates); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
}

func (h *Handler) deleteTrainingAPI(w http.ResponseWriter, r *http.Request, id int) {
	user, ok := h.requireCoach(w, r)
	if !ok {
		return
	}
	coach, err := h.coachService.GetCoachByUserID(user.ID)
	if err != nil || coach == nil {
		http.Error(w, "Coach profile not found", http.StatusBadRequest)
		return
	}
	existing, err := h.scheduleService.GetTrainingByID(id)
	if err != nil || existing == nil {
		http.Error(w, "Training not found", http.StatusNotFound)
		return
	}
	if existing.CoachID == nil || *existing.CoachID != coach.ID {
		http.Error(w, "Not your training", http.StatusForbidden)
		return
	}
	if err := h.scheduleService.DeleteTraining(id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
}

func (h *Handler) TrainingsFromTemplatesAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, ok := h.requireCoach(w, r)
	if !ok {
		return
	}
	coach, err := h.coachService.GetCoachByUserID(user.ID)
	if err != nil || coach == nil {
		http.Error(w, "Coach profile not found", http.StatusBadRequest)
		return
	}

	var req struct {
		WeeksCount int `json:"weeks_count"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if req.WeeksCount < 1 {
		req.WeeksCount = 4
	}
	if req.WeeksCount > 12 {
		req.WeeksCount = 12
	}

	weekStart := nextMonday(time.Now())
	created, err := h.scheduleService.CreateTrainingsFromTemplates(weekStart, coach.ID, user.ID, req.WeeksCount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	weekEnd := weekStart.AddDate(0, 0, req.WeeksCount*7-1)
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":            true,
		"created_count": created,
		"start":         weekStart.Format("2006-01-02"),
		"end":           weekEnd.Format("2006-01-02"),
		"weeks_count":   req.WeeksCount,
	})
}

func nextMonday(from time.Time) time.Time {
	daysUntilMonday := (8 - int(from.Weekday())) % 7
	if daysUntilMonday == 0 {
		daysUntilMonday = 7
	}
	return from.AddDate(0, 0, daysUntilMonday)
}

func parseLocalDateTime(date, hhmm string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02 15:04", date+" "+hhmm, time.Local)
}
