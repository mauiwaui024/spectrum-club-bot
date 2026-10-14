package web

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"spectrum-club-bot/internal/models"
)

// formatStudentDisplayName: Фамилия Имя (для алфавита по фамилии).
func formatStudentDisplayName(u *models.User) string {
	if u == nil {
		return "Неизвестный ученик"
	}
	last := strings.TrimSpace(u.LastName)
	first := strings.TrimSpace(u.FirstName)
	switch {
	case last != "" && first != "":
		return last + " " + first
	case last != "":
		return last
	case first != "":
		return first
	case strings.TrimSpace(u.Username) != "":
		return "@" + u.Username
	default:
		return "Неизвестный ученик"
	}
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *Handler) requireUser(w http.ResponseWriter, r *http.Request) (*models.User, bool) {
	userID, err := h.getUserIDFromRequest(r)
	if err != nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return nil, false
	}
	user, err := h.userService.GetByID(userID)
	if err != nil || user == nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return nil, false
	}
	return user, true
}

func (h *Handler) requireCoach(w http.ResponseWriter, r *http.Request) (*models.User, bool) {
	user, ok := h.requireUser(w, r)
	if !ok {
		return nil, false
	}
	if user.Role != "coach" {
		http.Error(w, "Coach access required", http.StatusForbidden)
		return nil, false
	}
	return user, true
}

func (h *Handler) setSessionCookie(w http.ResponseWriter, userID int64) error {
	sessionToken, err := h.createSession(userID)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   strings.ToLower(os.Getenv("ENVIRONMENT")) == "production",
		MaxAge:   60 * 60 * 24,
	})
	return nil
}
