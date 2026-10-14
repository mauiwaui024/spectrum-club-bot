package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"spectrum-club-bot/internal/models"
)

func TestGroupsAPI(t *testing.T) {
	h, _, _, _ := testHandler(1)
	token := withCoachSession(h, 1)
	req := httptest.NewRequest(http.MethodGet, "/api/groups", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
	rr := httptest.NewRecorder()
	h.GroupsAPI(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), `"Name":"НП"`) && !strings.Contains(rr.Body.String(), `"name":"НП"`) {
		// TrainingGroup JSON tags may vary
		if !strings.Contains(rr.Body.String(), "НП") {
			t.Fatalf("body=%s", rr.Body.String())
		}
	}
}

func TestCreateTrainingAPI_OK(t *testing.T) {
	h, _, _, sched := testHandler(1)
	token := withCoachSession(h, 1)
	body := `{"group_id":1,"date":"2030-06-01","start_time":"10:00","end_time":"11:30","description":"скалодром"}`
	req := httptest.NewRequest(http.MethodPost, "/api/trainings", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
	rr := httptest.NewRecorder()
	h.TrainingsAPI(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	if len(sched.created) != 1 {
		t.Fatalf("created=%d", len(sched.created))
	}
	if sched.created[0].Description != "скалодром" || sched.created[0].GroupID != 1 {
		t.Fatalf("training=%+v", sched.created[0])
	}
}

func TestCreateTrainingAPI_EndBeforeStart(t *testing.T) {
	h, _, _, _ := testHandler(1)
	token := withCoachSession(h, 1)
	body := `{"group_id":1,"date":"2030-06-01","start_time":"12:00","end_time":"11:00"}`
	req := httptest.NewRequest(http.MethodPost, "/api/trainings", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
	rr := httptest.NewRecorder()
	h.TrainingsAPI(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestPatchAndDeleteTrainingAPI(t *testing.T) {
	h, _, _, sched := testHandler(1)
	token := withCoachSession(h, 1)
	coachID := int64(5)
	tr := &models.TrainingSchedule{
		ID:          1,
		GroupID:     1,
		CoachID:     &coachID,
		Description: "old",
	}
	sched.trainings = map[int]*models.TrainingSchedule{1: tr}
	sched.created = []*models.TrainingSchedule{tr}

	patchBody := `{"description":"new desc"}`
	patchReq := httptest.NewRequest(http.MethodPatch, "/api/trainings/1", strings.NewReader(patchBody))
	patchReq.Header.Set("Content-Type", "application/json")
	patchReq.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
	patchRR := httptest.NewRecorder()
	h.TrainingByIDAPI(patchRR, patchReq)
	if patchRR.Code != http.StatusOK {
		t.Fatalf("patch status=%d body=%s", patchRR.Code, patchRR.Body.String())
	}
	if sched.trainings[1].Description != "new desc" {
		t.Fatalf("desc=%q", sched.trainings[1].Description)
	}

	delReq := httptest.NewRequest(http.MethodDelete, "/api/trainings/1", nil)
	delReq.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
	delRR := httptest.NewRecorder()
	h.TrainingByIDAPI(delRR, delReq)
	if delRR.Code != http.StatusOK {
		t.Fatalf("delete status=%d body=%s", delRR.Code, delRR.Body.String())
	}
	if len(sched.deleted) != 1 || sched.deleted[0] != 1 {
		t.Fatalf("deleted=%v", sched.deleted)
	}
}

func TestTrainingsFromTemplatesAPI(t *testing.T) {
	h, _, _, _ := testHandler(1)
	token := withCoachSession(h, 1)
	body := `{"weeks_count":2}`
	req := httptest.NewRequest(http.MethodPost, "/api/trainings/from-templates", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
	rr := httptest.NewRecorder()
	h.TrainingByIDAPI(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), `"created_count":3`) {
		t.Fatalf("body=%s", rr.Body.String())
	}
}
