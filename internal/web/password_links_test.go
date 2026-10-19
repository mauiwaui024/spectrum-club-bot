package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPasswordLinkSetsPasswordOnExistingUser(t *testing.T) {
	h, users, _, _ := testHandler(1)
	users.usersByID[2].IsActive = true
	access, refresh := withCoachSession(h, 1)

	listReq := httptest.NewRequest(http.MethodGet, "/api/users/without-password", nil)
	addAuthCookies(listReq, access, refresh)
	listRR := httptest.NewRecorder()
	h.UsersWithoutPasswordAPI(listRR, listReq)
	if listRR.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", listRR.Code, listRR.Body.String())
	}

	body, _ := json.Marshal(map[string]interface{}{"user_id": 2, "ttl_hours": 24})
	createReq := httptest.NewRequest(http.MethodPost, "/api/password-links", bytes.NewReader(body))
	createReq.Header.Set("Content-Type", "application/json")
	addAuthCookies(createReq, access, refresh)
	createRR := httptest.NewRecorder()
	h.PasswordLinksAPI(createRR, createReq)
	if createRR.Code != http.StatusOK {
		t.Fatalf("create status=%d body=%s", createRR.Code, createRR.Body.String())
	}
	var created map[string]interface{}
	_ = json.Unmarshal(createRR.Body.Bytes(), &created)
	url, _ := created["url"].(string)
	if url == "" || created["name"] != "Яковлева Анна" {
		t.Fatalf("bad link response: %s", createRR.Body.String())
	}

	token := url[len("https://example.test/set-password?token="):]
	acceptBody, _ := json.Marshal(map[string]string{"token": token, "password": "password123"})
	acceptReq := httptest.NewRequest(http.MethodPost, "/api/password-links/accept", bytes.NewReader(acceptBody))
	acceptReq.Header.Set("Content-Type", "application/json")
	acceptRR := httptest.NewRecorder()
	h.AcceptPasswordLinkAPI(acceptRR, acceptReq)
	if acceptRR.Code != http.StatusOK {
		t.Fatalf("accept status=%d body=%s", acceptRR.Code, acceptRR.Body.String())
	}

	again := httptest.NewRecorder()
	h.AcceptPasswordLinkAPI(again, httptest.NewRequest(http.MethodPost, "/api/password-links/accept", bytes.NewReader(acceptBody)))
	if again.Code != http.StatusForbidden {
		t.Fatalf("reuse status=%d body=%s", again.Code, again.Body.String())
	}
	if users.usersByID[2].ID != 2 {
		t.Fatal("user replaced")
	}
}
