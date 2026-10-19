package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func (h *Handler) UserByIDAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
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

	var req struct {
		Role     *string `json:"role"`
		IsActive *bool   `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Role == nil && req.IsActive == nil {
		http.Error(w, "role or is_active required", http.StatusBadRequest)
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
