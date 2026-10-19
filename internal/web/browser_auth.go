package web

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

const (
	accessCookieName  = "spectrum_access"
	refreshCookieName = "spectrum_refresh"
	legacySessionName = "spectrum_web_session"
	accessMaxAge      = 15 * 60
	refreshMaxAge     = 30 * 24 * 60 * 60
)

func (h *Handler) cookieSecure() bool {
	return strings.ToLower(os.Getenv("ENVIRONMENT")) == "production"
}

func (h *Handler) issueAuthCookies(w http.ResponseWriter, userID int64) error {
	if h.sessionService == nil {
		return errNoSessionService
	}
	access, refresh, err := h.sessionService.IssueSession(userID)
	if err != nil {
		return err
	}
	h.writeAuthCookies(w, access, refresh)
	return nil
}

func (h *Handler) writeAuthCookies(w http.ResponseWriter, access, refresh string) {
	secure := h.cookieSecure()
	http.SetCookie(w, &http.Cookie{
		Name:     accessCookieName,
		Value:    access,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
		MaxAge:   accessMaxAge,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    refresh,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
		MaxAge:   refreshMaxAge,
	})
	// Drop legacy in-memory session cookie if present.
	http.SetCookie(w, &http.Cookie{
		Name:     legacySessionName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
}

func (h *Handler) clearAuthCookies(w http.ResponseWriter) {
	secure := h.cookieSecure()
	for _, name := range []string{accessCookieName, refreshCookieName, legacySessionName} {
		http.SetCookie(w, &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Secure:   secure,
			MaxAge:   -1,
		})
	}
}

func cookieValue(r *http.Request, name string) string {
	c, err := r.Cookie(name)
	if err != nil || c == nil {
		return ""
	}
	return c.Value
}

func (h *Handler) BrowserLoginAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	req.Username = strings.TrimSpace(req.Username)

	user, err := h.userService.GetByUsername(req.Username)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}
	if !user.IsActive {
		http.Error(w, "Account deactivated", http.StatusForbidden)
		return
	}
	if !h.userService.VerifyBrowserPassword(user, req.Password) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := h.issueAuthCookies(w, user.ID); err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":       true,
		"user_id":  user.ID,
		"username": user.Username,
	})
}

func (h *Handler) BrowserSetPasswordAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := h.getUserIDFromRequest(w, r)
	if err != nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if err := h.userService.SetBrowserPassword(userID, req.Password); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
}

func (h *Handler) BrowserLogoutAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if h.sessionService != nil {
		_ = h.sessionService.RevokeRefresh(cookieValue(r, refreshCookieName))
	}
	h.clearAuthCookies(w)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
}

func (h *Handler) AuthStatusAPI(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromRequest(w, r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"authenticated": false,
		})
		return
	}

	user, userErr := h.userService.GetByID(userID)
	if userErr != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"authenticated": false,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"authenticated": true,
		"user_id":       user.ID,
		"username":      user.Username,
		"role":          user.Role,
		"is_coach":      user.Role == "coach",
		"pd_consent_ok": user.HasCurrentPDConsent(),
	})
}
