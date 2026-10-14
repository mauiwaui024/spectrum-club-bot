package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"spectrum-club-bot/internal/models"
)

func (h *Handler) TemplatesAPI(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireCoach(w, r); !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		templates, err := h.scheduleService.GetAllTemplates()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if templates == nil {
			templates = []models.WeekScheduleTemplate{}
		}
		for i := range templates {
			if g, err := h.groupService.GetGroupByID(templates[i].GroupID); err == nil && g != nil {
				templates[i].GroupName = g.Name
				templates[i].GroupCode = g.Code
			}
		}
		h.writeJSON(w, http.StatusOK, map[string]interface{}{"templates": templates})
	case http.MethodPost:
		h.createTemplateAPI(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) TemplateByIDAPI(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireCoach(w, r); !ok {
		return
	}
	rest := strings.TrimPrefix(r.URL.Path, "/api/templates/")
	rest = strings.Trim(rest, "/")
	parts := strings.Split(rest, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	id, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Error(w, "Invalid template id", http.StatusBadRequest)
		return
	}

	if len(parts) == 2 {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		switch parts[1] {
		case "activate":
			if err := h.scheduleService.ActivateTemplate(id); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			h.writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
		case "deactivate":
			if err := h.scheduleService.DeactivateTemplate(id); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			h.writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
		default:
			http.Error(w, "Not found", http.StatusNotFound)
		}
		return
	}

	switch r.Method {
	case http.MethodPatch:
		h.patchTemplateAPI(w, r, id)
	case http.MethodDelete:
		if err := h.scheduleService.DeleteTemplate(id); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) createTemplateAPI(w http.ResponseWriter, r *http.Request) {
	var req struct {
		GroupID     int    `json:"group_id"`
		DayOfWeek   int    `json:"day_of_week"`
		StartTime   string `json:"start_time"`
		EndTime     string `json:"end_time"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if req.GroupID <= 0 || req.DayOfWeek < 1 || req.DayOfWeek > 7 || req.StartTime == "" || req.EndTime == "" {
		http.Error(w, "group_id, day_of_week (1-7), start_time, end_time required", http.StatusBadRequest)
		return
	}
	if _, err := time.Parse("15:04", req.StartTime); err != nil {
		http.Error(w, "invalid start_time", http.StatusBadRequest)
		return
	}
	if _, err := time.Parse("15:04", req.EndTime); err != nil {
		http.Error(w, "invalid end_time", http.StatusBadRequest)
		return
	}
	tpl := &models.WeekScheduleTemplate{
		GroupID:     req.GroupID,
		DayOfWeek:   req.DayOfWeek,
		StartTime:   req.StartTime + ":00",
		EndTime:     req.EndTime + ":00",
		Description: req.Description,
		IsActive:    true,
	}
	if err := h.scheduleService.CreateTemplate(tpl); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "template": tpl})
}

func (h *Handler) patchTemplateAPI(w http.ResponseWriter, r *http.Request, id int) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	updates := map[string]interface{}{}
	if v, ok := req["group_id"].(float64); ok {
		updates["group_id"] = int(v)
	}
	if v, ok := req["day_of_week"].(float64); ok {
		updates["day_of_week"] = int(v)
	}
	if v, ok := req["description"].(string); ok {
		updates["description"] = v
	}
	if v, ok := req["is_active"].(bool); ok {
		updates["is_active"] = v
	}
	if v, ok := req["start_time"].(string); ok && v != "" {
		if _, err := time.Parse("15:04", v); err == nil {
			updates["start_time"] = v + ":00"
		}
	}
	if v, ok := req["end_time"].(string); ok && v != "" {
		if _, err := time.Parse("15:04", v); err == nil {
			updates["end_time"] = v + ":00"
		}
	}
	if len(updates) == 0 {
		http.Error(w, "No updates", http.StatusBadRequest)
		return
	}
	if err := h.scheduleService.UpdateTemplate(id, updates); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
}

func (h *Handler) createGroupAPI(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Code        string `json:"code"`
		AgeMin      int    `json:"age_min"`
		AgeMax      *int   `json:"age_max"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	g := &models.TrainingGroup{
		Name: req.Name, Code: req.Code, AgeMin: req.AgeMin, AgeMax: req.AgeMax, Description: req.Description,
	}
	if err := h.groupService.CreateGroup(g); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "group": g})
}

func (h *Handler) GroupByIDAPI(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireCoach(w, r); !ok {
		return
	}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/groups/")
	idStr = strings.Trim(idStr, "/")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid group id", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodPatch:
		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if err := h.groupService.UpdateGroup(id, req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
	case http.MethodDelete:
		if err := h.groupService.DeleteGroup(id); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
