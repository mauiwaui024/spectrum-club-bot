package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"spectrum-club-bot/internal/service"
)

func TestRegisterStudentByCoachAPI(t *testing.T) {
	h, _, _, _ := testHandler(1)
	access, refresh := withCoachSession(h, 1)

	req := httptest.NewRequest(http.MethodPost, "/api/trainings/register-student", strings.NewReader(`{"training_id":3,"student_id":20}`))
	req.Header.Set("Content-Type", "application/json")
	addAuthCookies(req, access, refresh)
	rr := httptest.NewRecorder()
	h.RegisterStudentByCoachAPI(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}

	h.attendanceService.(*mockAttendanceService).coachSignupErr = service.ErrNoSubscription
	again := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/api/trainings/register-student", strings.NewReader(`{"training_id":3,"student_id":20}`))
	addAuthCookies(req2, access, refresh)
	h.RegisterStudentByCoachAPI(again, req2)
	if again.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", again.Code, again.Body.String())
	}
}
