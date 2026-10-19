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

type invitePayload struct {
	Role      string `json:"role"`
	Exp       int64  `json:"exp"`
	JTI       string `json:"jti"`
	CreatedBy int64  `json:"created_by"`
}

func (h *Handler) CreateInviteAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	coach, ok := h.requireCoach(w, r)
	if !ok {
		return
	}

	var req struct {
		Role     string `json:"role"`
		TTLHours int    `json:"ttl_hours"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	role := strings.TrimSpace(req.Role)
	if role != "student" && role != "coach" {
		http.Error(w, "role must be student or coach", http.StatusBadRequest)
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
		http.Error(w, "Failed to create invite", http.StatusInternalServerError)
		return
	}

	payload := invitePayload{
		Role:      role,
		Exp:       time.Now().Add(time.Duration(ttl) * time.Hour).Unix(),
		JTI:       hex.EncodeToString(jtiBytes),
		CreatedBy: coach.ID,
	}
	token, err := h.signInvite(payload)
	if err != nil {
		http.Error(w, "Failed to sign invite", http.StatusInternalServerError)
		return
	}

	base := h.baseURL
	if base == "" {
		port := os.Getenv("HTTP_PORT")
		if port == "" {
			port = "8080"
		}
		base = "http://localhost:" + port
	}
	url := fmt.Sprintf("%s/register?token=%s", base, token)

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"token":      token,
		"url":        url,
		"role":       role,
		"expires_at": time.Unix(payload.Exp, 0).UTC().Format(time.RFC3339),
	})
}

func (h *Handler) RegisterInviteAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Token         string `json:"token"`
		Username      string `json:"username"`
		Password      string `json:"password"`
		FirstName     string `json:"first_name"`
		LastName      string `json:"last_name"`
		PDConsent     bool   `json:"pd_consent"`
		PolicyVersion string `json:"policy_version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	payload, err := h.verifyInvite(strings.TrimSpace(req.Token))
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	if h.isInviteUsed(payload.JTI) {
		http.Error(w, "invite already used", http.StatusForbidden)
		return
	}

	if !req.PDConsent || req.PolicyVersion != models.PDPolicyVersion {
		http.Error(w, "Нужно согласие на обработку персональных данных", http.StatusBadRequest)
		return
	}

	user, err := h.userService.CreateWebUser(req.Username, req.Password, req.FirstName, req.LastName, payload.Role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.userService.RecordPDConsent(user.ID, models.PDPolicyVersion); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.markInviteUsed(payload.JTI, time.Unix(payload.Exp, 0))

	if err := h.issueAuthCookies(w, user.ID); err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":       true,
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
	})
}

func (h *Handler) ValidateInviteAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	payload, err := h.verifyInvite(token)
	if err != nil {
		h.writeJSON(w, http.StatusOK, map[string]interface{}{"ok": false, "error": err.Error()})
		return
	}
	if h.isInviteUsed(payload.JTI) {
		h.writeJSON(w, http.StatusOK, map[string]interface{}{"ok": false, "error": "invite already used"})
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"role":       payload.Role,
		"expires_at": time.Unix(payload.Exp, 0).UTC().Format(time.RFC3339),
	})
}

func (h *Handler) signInvite(payload invitePayload) (string, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, []byte(h.inviteSecret))
	_, _ = mac.Write(raw)
	sig := mac.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(raw) + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

func (h *Handler) verifyInvite(token string) (*invitePayload, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid invite token")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid invite token")
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid invite token")
	}
	mac := hmac.New(sha256.New, []byte(h.inviteSecret))
	_, _ = mac.Write(raw)
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return nil, fmt.Errorf("invalid invite signature")
	}
	var payload invitePayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("invalid invite payload")
	}
	if payload.Role != "student" && payload.Role != "coach" {
		return nil, fmt.Errorf("invalid invite role")
	}
	if time.Now().Unix() > payload.Exp {
		return nil, fmt.Errorf("invite expired")
	}
	return &payload, nil
}

func (h *Handler) isInviteUsed(jti string) bool {
	h.usedInviteMu.Lock()
	defer h.usedInviteMu.Unlock()
	h.cleanupUsedInvitesLocked()
	_, used := h.usedInviteJTIs[jti]
	return used
}

func (h *Handler) markInviteUsed(jti string, exp time.Time) {
	h.usedInviteMu.Lock()
	defer h.usedInviteMu.Unlock()
	h.usedInviteJTIs[jti] = exp
	h.cleanupUsedInvitesLocked()
}

func (h *Handler) cleanupUsedInvitesLocked() {
	now := time.Now()
	for jti, exp := range h.usedInviteJTIs {
		if now.After(exp) {
			delete(h.usedInviteJTIs, jti)
		}
	}
}
