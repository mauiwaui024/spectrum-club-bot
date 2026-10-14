package web

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"spectrum-club-bot/internal/models"
	"spectrum-club-bot/internal/service"
)

func (h *Handler) SetPlanService(s service.PlanService) {
	h.planService = s
}

type planBody struct {
	Title            string            `json:"title"`
	Description      string            `json:"description"`
	Items            []models.PlanItem `json:"items"`
	SourceTemplateID *int              `json:"source_template_id"`
}

func (h *Handler) PlanTemplatesAPI(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireCoach(w, r); !ok {
		return
	}
	if h.planService == nil {
		http.Error(w, "Plans are not configured", http.StatusInternalServerError)
		return
	}
	switch r.Method {
	case http.MethodGet:
		list, err := h.planService.ListTemplates()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		h.writeJSON(w, http.StatusOK, map[string]interface{}{"templates": list})
	case http.MethodPost:
		body, items, ok := h.readPlanBody(w, r, true)
		if !ok {
			return
		}
		created, err := h.planService.CreateTemplate(body.Title, body.Description, items)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		h.writeJSON(w, http.StatusCreated, map[string]interface{}{"template": created})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) PlanTemplateByIDAPI(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireCoach(w, r); !ok {
		return
	}
	if h.planService == nil {
		http.Error(w, "Plans are not configured", http.StatusInternalServerError)
		return
	}
	id, err := strconv.Atoi(strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/plan-templates/"), "/"))
	if err != nil || id <= 0 {
		http.Error(w, "Invalid template id", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodPatch:
		body, items, ok := h.readPlanBody(w, r, true)
		if !ok {
			return
		}
		if err := h.planService.UpdateTemplate(id, body.Title, body.Description, items); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "Template not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tpl, err := h.planService.GetTemplate(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		h.writeJSON(w, http.StatusOK, map[string]interface{}{"template": tpl})
	case http.MethodDelete:
		if err := h.planService.DeleteTemplate(id); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "Template not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		h.writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) trainingPlanAPI(w http.ResponseWriter, r *http.Request, trainingID int) {
	if h.planService == nil {
		http.Error(w, "Plans are not configured", http.StatusInternalServerError)
		return
	}
	switch r.Method {
	case http.MethodGet:
		if _, ok := h.requireUser(w, r); !ok {
			return
		}
		plan, err := h.planService.GetTrainingPlan(trainingID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		h.writeJSON(w, http.StatusOK, map[string]interface{}{"plan": plan})
	case http.MethodPut:
		if _, ok := h.requireCoach(w, r); !ok {
			return
		}
		if _, err := h.scheduleService.GetTrainingByID(trainingID); err != nil {
			http.Error(w, "Training not found", http.StatusNotFound)
			return
		}
		body, items, ok := h.readPlanBody(w, r, false)
		if !ok {
			return
		}
		if body.SourceTemplateID != nil {
			tpl, err := h.planService.GetTemplate(*body.SourceTemplateID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if tpl == nil {
				http.Error(w, "Template not found", http.StatusBadRequest)
				return
			}
		}
		saved, err := h.planService.SaveTrainingPlan(trainingID, body.SourceTemplateID, body.Description, items)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		h.writeJSON(w, http.StatusOK, map[string]interface{}{"plan": saved})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) readPlanBody(w http.ResponseWriter, r *http.Request, requireTitle bool) (planBody, []models.PlanItem, bool) {
	var body planBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return body, nil, false
	}
	body.Title = strings.TrimSpace(body.Title)
	body.Description = strings.TrimSpace(body.Description)
	if requireTitle && body.Title == "" {
		http.Error(w, "Укажите название", http.StatusBadRequest)
		return body, nil, false
	}
	if len([]rune(body.Title)) > 255 || len([]rune(body.Description)) > 4000 {
		http.Error(w, "Слишком длинный текст", http.StatusBadRequest)
		return body, nil, false
	}
	items := make([]models.PlanItem, 0, len(body.Items))
	for _, item := range body.Items {
		sets, ok := cleanPlanCount(w, item.Sets)
		if !ok {
			return body, nil, false
		}
		reps, ok := cleanPlanCount(w, item.Reps)
		if !ok {
			return body, nil, false
		}
		rest, ok := cleanPlanCount(w, item.RestSeconds)
		if !ok {
			return body, nil, false
		}
		text := strings.TrimSpace(item.Text)
		var exerciseID *int
		if item.ExerciseID != nil && *item.ExerciseID > 0 {
			ex, err := h.planService.GetExercise(*item.ExerciseID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return body, nil, false
			}
			if ex == nil {
				http.Error(w, "Упражнение не найдено", http.StatusBadRequest)
				return body, nil, false
			}
			text = ex.Name
			exerciseID = item.ExerciseID
		}
		if text == "" {
			continue
		}
		if len([]rune(text)) > 500 {
			http.Error(w, "Слишком длинное упражнение", http.StatusBadRequest)
			return body, nil, false
		}
		items = append(items, models.PlanItem{
			Text:        text,
			Sets:        sets,
			Reps:        reps,
			RestSeconds: rest,
			ExerciseID:  exerciseID,
		})
	}
	if len(items) > 80 {
		http.Error(w, "Слишком много упражнений", http.StatusBadRequest)
		return body, nil, false
	}
	return body, items, true
}

func cleanPlanCount(w http.ResponseWriter, n *int) (*int, bool) {
	if n == nil || *n <= 0 {
		return nil, true
	}
	if *n > 999 {
		http.Error(w, "Слишком большое число", http.StatusBadRequest)
		return nil, false
	}
	return n, true
}

func (h *Handler) trainingPlanOrNil(trainingID int) *models.TrainingPlan {
	if h.planService == nil {
		return nil
	}
	plan, err := h.planService.GetTrainingPlan(trainingID)
	if err != nil {
		return nil
	}
	return plan
}
