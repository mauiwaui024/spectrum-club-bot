package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestStudentsAPI_DisplayNameAndStudentID(t *testing.T) {
	h, _, _, _ := testHandler(1)
	access, refresh := withCoachSession(h, 1)
	req := httptest.NewRequest(http.MethodGet, "/api/students", nil)
	addAuthCookies(req, access, refresh)
	rr := httptest.NewRecorder()
	h.StudentsAPI(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	if !strings.Contains(body, `"display_name":"Яковлева Анна"`) {
		t.Fatalf("missing display_name: %s", body)
	}
	if !strings.Contains(body, `"student_id":20`) {
		t.Fatalf("missing student_id: %s", body)
	}
}

func TestCreateSubscriptionAPI_OK(t *testing.T) {
	h, _, subs, _ := testHandler(1)
	access, refresh := withCoachSession(h, 1)
	end := time.Now().AddDate(0, 0, 30).Format("2006-01-02")
	body := `{"student_user_id":2,"lessons":8,"end_date":"` + end + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/subscriptions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	addAuthCookies(req, access, refresh)
	rr := httptest.NewRecorder()
	h.SubscriptionsAPI(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	if len(subs.created) != 1 || subs.created[0].studentID != 20 || subs.created[0].lessons != 8 {
		t.Fatalf("created=%+v", subs.created)
	}
}

func TestCreateSubscriptionAPI_LegacyType(t *testing.T) {
	h, _, subs, _ := testHandler(1)
	access, refresh := withCoachSession(h, 1)
	end := time.Now().AddDate(0, 0, 14).Format("2006-01-02")
	body := `{"student_user_id":2,"type":"12","end_date":"` + end + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/subscriptions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	addAuthCookies(req, access, refresh)
	rr := httptest.NewRecorder()
	h.SubscriptionsAPI(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	if len(subs.created) != 1 || subs.created[0].lessons != 12 {
		t.Fatalf("created=%+v", subs.created)
	}
}

func TestCreateSubscriptionAPI_InvalidLessons(t *testing.T) {
	h, _, _, _ := testHandler(1)
	access, refresh := withCoachSession(h, 1)
	body := `{"student_user_id":2,"lessons":5,"end_date":"2030-01-01"}`
	req := httptest.NewRequest(http.MethodPost, "/api/subscriptions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	addAuthCookies(req, access, refresh)
	rr := httptest.NewRecorder()
	h.SubscriptionsAPI(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestDeleteSubscriptionAPI(t *testing.T) {
	h, _, subs, _ := testHandler(1)
	access, refresh := withCoachSession(h, 1)
	req := httptest.NewRequest(http.MethodDelete, "/api/subscriptions/99", nil)
	addAuthCookies(req, access, refresh)
	rr := httptest.NewRecorder()
	h.SubscriptionByIDAPI(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	if len(subs.deleted) != 1 || subs.deleted[0] != 99 {
		t.Fatalf("deleted=%v", subs.deleted)
	}
}
