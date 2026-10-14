package web

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

func (h *Handler) ExercisesAPI(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireCoach(w, r); !ok {
		return
	}
	if h.planService == nil {
		http.Error(w, "Plans are not configured", http.StatusInternalServerError)
		return
	}
	switch r.Method {
	case http.MethodGet:
		list, err := h.planService.ListExercises()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		h.writeJSON(w, http.StatusOK, map[string]interface{}{"exercises": list})
	case http.MethodPost:
		name, description, ok := h.readExerciseBody(w, r)
		if !ok {
			return
		}
		created, err := h.planService.CreateExercise(name, description)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.writeJSON(w, http.StatusCreated, map[string]interface{}{"exercise": created})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) ExerciseByIDAPI(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireCoach(w, r); !ok {
		return
	}
	if h.planService == nil {
		http.Error(w, "Plans are not configured", http.StatusInternalServerError)
		return
	}
	id, err := strconv.Atoi(strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/exercises/"), "/"))
	if err != nil || id <= 0 {
		http.Error(w, "Invalid exercise id", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodPatch:
		name, description, ok := h.readExerciseBody(w, r)
		if !ok {
			return
		}
		if err := h.planService.UpdateExercise(id, name, description); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "Exercise not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		ex, err := h.planService.GetExercise(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		h.writeJSON(w, http.StatusOK, map[string]interface{}{"exercise": ex})
	case http.MethodDelete:
		if err := h.planService.DeleteExercise(id); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "Exercise not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) readExerciseBody(w http.ResponseWriter, r *http.Request) (string, string, bool) {
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return "", "", false
	}
	name := strings.TrimSpace(body.Name)
	description := strings.TrimSpace(body.Description)
	if name == "" {
		http.Error(w, "Укажите название", http.StatusBadRequest)
		return "", "", false
	}
	if len([]rune(name)) > 255 || len([]rune(description)) > 4000 {
		http.Error(w, "Слишком длинный текст", http.StatusBadRequest)
		return "", "", false
	}
	return name, description, true
}
