package web

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"spectrum-club-bot/internal/models"
)

type passwordLinkPayload struct {
	Purpose string `json:"purpose"`
	UserID  int64  `json:"user_id"`
	Exp     int64  `json:"exp"`
	JTI     string `json:"jti"`
}

func (h *Handler) UsersWithoutPasswordAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := h.requireCoach(w, r); !ok {
		return
	}
	users, err := h.userService.ListUsersWithoutPassword()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	out := make([]map[string]interface{}, 0, len(users))
	for _, u := range users {
		out = append(out, map[string]interface{}{
			"id":         u.ID,
			"first_name": u.FirstName,
			"last_name":  u.LastName,
			"username":   u.Username,
			"role":       u.Role,
			"name":       formatStudentDisplayName(u),
		})
	}
	h.writeJSON(w, http.StatusOK, map[string]interface{}{"users": out})
}

func (h *Handler) PasswordLinksAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := h.requireCoach(w, r); !ok {
		return
	}
	var req struct {
		UserID   int64 `json:"user_id"`
		TTLHours int   `json:"ttl_hours"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	user, err := h.userService.GetByID(req.UserID)
	if err != nil || user == nil {
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}
	if !user.IsActive {
		http.Error(w, "Вход для этого человека закрыт", http.StatusBadRequest)
		return
	}
	if userHasPassword(user) {
		http.Error(w, "Пароль уже задан", http.StatusBadRequest)
		return
	}
	ttl := req.TTLHours
	if ttl <= 0 {
		ttl = 72
	}
	if ttl > 24*14 {
		ttl = 24 * 14
	}
	jtiBytes := make([]byte, 16)
	if _, err := rand.Read(jtiBytes); err != nil {
		http.Error(w, "Failed to create link", http.StatusInternalServerError)
		return
	}
	payload := passwordLinkPayload{
		Purpose: "set_password",
		UserID:  user.ID,
		Exp:     time.Now().Add(time.Duration(ttl) * time.Hour).Unix(),
		JTI:     hex.EncodeToString(jtiBytes),
	}
	token, err := h.signPasswordLink(payload)
	if err != nil {
		http.Error(w, "Failed to sign link", http.StatusInternalServerError)
		return
	}
	url := fmt.Sprintf("%s/set-password?token=%s", h.publicBaseURL(), token)
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"url":        url,
		"expires_at": time.Unix(payload.Exp, 0).UTC().Format(time.RFC3339),
		"name":       formatStudentDisplayName(user),
	})
}

func (h *Handler) ValidatePasswordLinkAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	payload, user, err := h.openPasswordLink(r.URL.Query().Get("token"))
	if err != nil {
		h.writeJSON(w, http.StatusOK, map[string]interface{}{"ok": false, "error": err.Error()})
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":          true,
		"name":        formatStudentDisplayName(user),
		"username":    strings.TrimSpace(user.Username),
		"needs_login": strings.TrimSpace(user.Username) == "",
		"expires_at":  time.Unix(payload.Exp, 0).UTC().Format(time.RFC3339),
	})
}

func (h *Handler) AcceptPasswordLinkAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Token    string `json:"token"`
		Password string `json:"password"`
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	payload, user, err := h.openPasswordLink(req.Token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	if strings.TrimSpace(user.Username) == "" {
		if err := h.userService.SetUsernameIfEmpty(user.ID, req.Username); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	if err := h.userService.SetBrowserPassword(user.ID, req.Password); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.markInviteUsed(payload.JTI, time.Unix(payload.Exp, 0))
	if err := h.issueAuthCookies(w, user.ID); err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "user_id": user.ID})
}

func (h *Handler) openPasswordLink(token string) (*passwordLinkPayload, *models.User, error) {
	payload, err := h.verifyPasswordLink(strings.TrimSpace(token))
	if err != nil {
		return nil, nil, err
	}
	if h.isInviteUsed(payload.JTI) {
		return nil, nil, fmt.Errorf("ссылка уже использована")
	}
	user, err := h.userService.GetByID(payload.UserID)
	if err != nil || user == nil {
		return nil, nil, fmt.Errorf("пользователь не найден")
	}
	if !user.IsActive {
		return nil, nil, fmt.Errorf("вход для этого человека закрыт")
	}
	if userHasPassword(user) {
		return nil, nil, fmt.Errorf("пароль уже задан")
	}
	return payload, user, nil
}

func userHasPassword(u *models.User) bool {
	return u != nil && u.PasswordHash != nil && strings.TrimSpace(*u.PasswordHash) != ""
}

func (h *Handler) publicBaseURL() string {
	if h.baseURL != "" {
		return h.baseURL
	}
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}
	return "http://localhost:" + port
}

func (h *Handler) signPasswordLink(payload passwordLinkPayload) (string, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, []byte(h.inviteSecret))
	_, _ = mac.Write(raw)
	sig := mac.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(raw) + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

func (h *Handler) verifyPasswordLink(token string) (*passwordLinkPayload, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return nil, fmt.Errorf("ссылка недействительна")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("ссылка недействительна")
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("ссылка недействительна")
	}
	mac := hmac.New(sha256.New, []byte(h.inviteSecret))
	_, _ = mac.Write(raw)
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return nil, fmt.Errorf("ссылка недействительна")
	}
	var payload passwordLinkPayload
	if err := json.Unmarshal(raw, &payload); err != nil || payload.Purpose != "set_password" || payload.UserID <= 0 {
		return nil, fmt.Errorf("ссылка недействительна")
	}
	if time.Now().Unix() > payload.Exp {
		return nil, fmt.Errorf("срок ссылки истёк")
	}
	return &payload, nil
}
