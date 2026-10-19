package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"spectrum-club-bot/internal/models"
)

func TestInviteSignAndVerify(t *testing.T) {
	h := &Handler{inviteSecret: "secret", usedInviteJTIs: map[string]time.Time{}}
	payload := invitePayload{
		Role:      "student",
		Exp:       time.Now().Add(time.Hour).Unix(),
		JTI:       "abc123",
		CreatedBy: 1,
	}
	token, err := h.signInvite(payload)
	if err != nil {
		t.Fatal(err)
	}
	got, err := h.verifyInvite(token)
	if err != nil {
		t.Fatal(err)
	}
	if got.Role != "student" || got.JTI != "abc123" || got.CreatedBy != 1 {
		t.Fatalf("unexpected payload: %+v", got)
	}
}

func TestInviteVerify_Tampered(t *testing.T) {
	h := &Handler{inviteSecret: "secret", usedInviteJTIs: map[string]time.Time{}}
	token, _ := h.signInvite(invitePayload{
		Role: "student", Exp: time.Now().Add(time.Hour).Unix(), JTI: "x", CreatedBy: 1,
	})
	parts := strings.Split(token, ".")
	bad := parts[0] + ".deadbeef"
	if _, err := h.verifyInvite(bad); err == nil {
		t.Fatal("expected signature error")
	}
}

func TestInviteVerify_Expired(t *testing.T) {
	h := &Handler{inviteSecret: "secret", usedInviteJTIs: map[string]time.Time{}}
	token, _ := h.signInvite(invitePayload{
		Role: "coach", Exp: time.Now().Add(-time.Hour).Unix(), JTI: "y", CreatedBy: 1,
	})
	if _, err := h.verifyInvite(token); err == nil {
		t.Fatal("expected expired error")
	}
}

func TestCreateInviteAPI_CoachOnly(t *testing.T) {
	h, _, _, _ := testHandler(1)
	stuAccess, stuRefresh, _ := h.sessionService.IssueSession(2)
	req := httptest.NewRequest(http.MethodPost, "/api/invites", strings.NewReader(`{"role":"student"}`))
	req.Header.Set("Content-Type", "application/json")
	addAuthCookies(req, stuAccess, stuRefresh)
	rr := httptest.NewRecorder()
	h.CreateInviteAPI(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestCreateInviteAndRegisterFlow(t *testing.T) {
	h, users, _, _ := testHandler(1)
	access, refresh := withCoachSession(h, 1)

	req := httptest.NewRequest(http.MethodPost, "/api/invites", strings.NewReader(`{"role":"student","ttl_hours":24}`))
	req.Header.Set("Content-Type", "application/json")
	addAuthCookies(req, access, refresh)
	rr := httptest.NewRecorder()
	h.CreateInviteAPI(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("create invite status=%d body=%s", rr.Code, rr.Body.String())
	}
	var inviteResp map[string]interface{}
	_ = json.Unmarshal(rr.Body.Bytes(), &inviteResp)
	inviteToken, _ := inviteResp["token"].(string)
	if inviteToken == "" {
		t.Fatal("empty invite token")
	}
	url, _ := inviteResp["url"].(string)
	if !strings.Contains(url, "/register?token=") {
		t.Fatalf("bad url: %s", url)
	}

	body, _ := json.Marshal(map[string]interface{}{
		"token":          inviteToken,
		"username":       "newkid",
		"password":       "password123",
		"first_name":     "Никита",
		"last_name":      "Сидоров",
		"pd_consent":     true,
		"policy_version": models.PDPolicyVersion,
	})
	regReq := httptest.NewRequest(http.MethodPost, "/api/register-invite", bytes.NewReader(body))
	regReq.Header.Set("Content-Type", "application/json")
	regRR := httptest.NewRecorder()
	h.RegisterInviteAPI(regRR, regReq)
	if regRR.Code != http.StatusOK {
		t.Fatalf("register status=%d body=%s", regRR.Code, regRR.Body.String())
	}
	if len(users.created) != 1 || users.created[0].Username != "newkid" {
		t.Fatalf("user not created: %+v", users.created)
	}
	// reuse invite → forbidden
	regRR2 := httptest.NewRecorder()
	regReq2 := httptest.NewRequest(http.MethodPost, "/api/register-invite", bytes.NewReader(body))
	regReq2.Header.Set("Content-Type", "application/json")
	h.RegisterInviteAPI(regRR2, regReq2)
	if regRR2.Code != http.StatusForbidden {
		t.Fatalf("reuse expected 403, got %d", regRR2.Code)
	}
}

func TestValidateInviteAPI(t *testing.T) {
	h, _, _, _ := testHandler(1)
	token, err := h.signInvite(invitePayload{
		Role: "student", Exp: time.Now().Add(time.Hour).Unix(), JTI: "val1", CreatedBy: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/invites/validate?token="+token, nil)
	rr := httptest.NewRecorder()
	h.ValidateInviteAPI(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	var resp map[string]interface{}
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp["ok"] != true || resp["role"] != "student" {
		t.Fatalf("resp=%v", resp)
	}
}

func TestFormatStudentDisplayName(t *testing.T) {
	cases := []struct {
		u    *models.User
		want string
	}{
		{&models.User{LastName: "Яковлева", FirstName: "Анна"}, "Яковлева Анна"},
		{&models.User{FirstName: "Анна"}, "Анна"},
		{&models.User{LastName: "Яковлева"}, "Яковлева"},
		{&models.User{Username: "nick"}, "@nick"},
		{&models.User{}, "Неизвестный ученик"},
		{nil, "Неизвестный ученик"},
	}
	for _, c := range cases {
		if got := formatStudentDisplayName(c.u); got != c.want {
			t.Fatalf("got %q want %q for %+v", got, c.want, c.u)
		}
	}
}
