package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func (h *Handler) UserByIDAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch && r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	actor, ok := h.requireCoach(w, r)
	if !ok {
		return
	}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/users/")
	idStr = strings.Trim(idStr, "/")
	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || userID <= 0 {
		http.Error(w, "Invalid user id", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodDelete {
		h.deleteStudentUser(w, actor.ID, userID)
		return
	}

	var req struct {
		Role      *string `json:"role"`
		IsActive  *bool   `json:"is_active"`
		FirstName *string `json:"first_name"`
		LastName  *string `json:"last_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Role == nil && req.IsActive == nil && req.FirstName == nil && req.LastName == nil {
		http.Error(w, "role, is_active, or name required", http.StatusBadRequest)
		return
	}

	target, err := h.userService.GetByID(userID)
	if err != nil || target == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if req.IsActive != nil {
		if userID == actor.ID && !*req.IsActive {
			http.Error(w, "cannot deactivate yourself", http.StatusBadRequest)
			return
		}
		if err := h.userService.SetActive(userID, *req.IsActive); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if !*req.IsActive && h.sessionService != nil {
			_ = h.sessionService.RevokeAllForUser(userID)
		}
	}

	if req.FirstName != nil || req.LastName != nil {
		if req.FirstName == nil || req.LastName == nil {
			http.Error(w, "укажите имя и фамилию", http.StatusBadRequest)
			return
		}
		first := strings.TrimSpace(*req.FirstName)
		last := strings.TrimSpace(*req.LastName)
		if first == "" || last == "" {
			http.Error(w, "укажите имя и фамилию", http.StatusBadRequest)
			return
		}
		if err := h.userService.UpdateProfile(userID, &first, &last); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	if req.Role != nil {
		role := strings.TrimSpace(*req.Role)
		if role != "student" && role != "coach" {
			http.Error(w, "role must be student or coach", http.StatusBadRequest)
			return
		}
		if err := h.userService.SetRoleByUserID(userID, role); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if role == "coach" {
			if err := h.userService.EnsureCoachProfile(userID); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
}

func (h *Handler) deleteStudentUser(w http.ResponseWriter, actorID, userID int64) {
	if userID == actorID {
		http.Error(w, "нельзя удалить себя", http.StatusBadRequest)
		return
	}
	target, err := h.userService.GetByID(userID)
	if err != nil || target == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	if target.Role != "student" {
		http.Error(w, "удалить можно только ученика", http.StatusBadRequest)
		return
	}
	if h.sessionService != nil {
		_ = h.sessionService.RevokeAllForUser(userID)
	}
	if err := h.userService.DeleteUser(userID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
}
