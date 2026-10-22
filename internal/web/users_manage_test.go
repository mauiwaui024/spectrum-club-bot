package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDeleteStudentUser(t *testing.T) {
	h, users, _, _ := testHandler(1)
	access, refresh := withCoachSession(h, 1)

	self := httptest.NewRequest(http.MethodDelete, "/api/users/1", nil)
	addAuthCookies(self, access, refresh)
	selfRR := httptest.NewRecorder()
	h.UserByIDAPI(selfRR, self)
	if selfRR.Code != http.StatusBadRequest {
		t.Fatalf("self status=%d body=%s", selfRR.Code, selfRR.Body.String())
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/users/2", nil)
	addAuthCookies(req, access, refresh)
	rr := httptest.NewRecorder()
	h.UserByIDAPI(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("delete status=%d body=%s", rr.Code, rr.Body.String())
	}
	if _, ok := users.usersByID[2]; ok {
		t.Fatal("student still present")
	}
}

func TestPatchStudentName(t *testing.T) {
	h, users, _, _ := testHandler(1)
	access, refresh := withCoachSession(h, 1)
	req := httptest.NewRequest(http.MethodPatch, "/api/users/2", strings.NewReader(`{"first_name":"Мария","last_name":"Иванова"}`))
	req.Header.Set("Content-Type", "application/json")
	addAuthCookies(req, access, refresh)
	rr := httptest.NewRecorder()
	h.UserByIDAPI(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	if users.usersByID[2].FirstName != "Мария" || users.usersByID[2].LastName != "Иванова" {
		t.Fatalf("name=%s %s", users.usersByID[2].FirstName, users.usersByID[2].LastName)
	}
}
