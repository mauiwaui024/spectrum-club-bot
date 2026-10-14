package web

import (
	"encoding/json"
	"net/http"

	"spectrum-club-bot/internal/models"
)

func (h *Handler) PDConsentAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, ok := h.requireUser(w, r)
	if !ok {
		return
	}
	var body struct {
		PolicyVersion string `json:"policy_version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if body.PolicyVersion != models.PDPolicyVersion {
		http.Error(w, "Нужно согласие на обработку персональных данных", http.StatusBadRequest)
		return
	}
	if err := h.userService.RecordPDConsent(user.ID, models.PDPolicyVersion); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "pd_consent_ok": true})
}
