package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (h *Handler) StudentsAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := h.requireCoach(w, r); !ok {
		return
	}
	students, err := h.userService.GetAllStudents()
	if err != nil {
		http.Error(w, "Failed to load students", http.StatusInternalServerError)
		return
	}

	filter := strings.TrimSpace(r.URL.Query().Get("filter"))
	if filter == "" {
		filter = "all"
	}
	days := 14
	if d := r.URL.Query().Get("days"); d != "" {
		if n, err := strconv.Atoi(d); err == nil && n > 0 {
			days = n
		}
	}
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))

	type studentDTO struct {
		UserID      int64  `json:"user_id"`
		StudentID   int64  `json:"student_id"`
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
		IsActive    bool   `json:"is_active"`
		Role        string `json:"role"`
	}
	out := make([]studentDTO, 0, len(students))
	now := time.Now()
	expCutoff := now.AddDate(0, 0, days)
	for _, u := range students {
		if q != "" {
			hay := strings.ToLower(u.LastName + " " + u.FirstName + " " + u.Username)
			if !strings.Contains(hay, q) {
				continue
			}
		}
		st, _ := h.studentService.GetStudentByUserID(u.ID)
		var studentID int64
		if st != nil {
			studentID = st.ID
		}
		activeSub, _ := h.subscriptionService.GetActiveSubscription(studentID)
		switch filter {
		case "no_subscription":
			if activeSub != nil {
				continue
			}
		case "expiring":
			if activeSub == nil || activeSub.EndDate.After(expCutoff) {
				continue
			}
		case "all":
		default:
			http.Error(w, "filter must be all|no_subscription|expiring", http.StatusBadRequest)
			return
		}
		out = append(out, studentDTO{
			UserID:      u.ID,
			StudentID:   studentID,
			FirstName:   u.FirstName,
			LastName:    u.LastName,
			Username:    u.Username,
			DisplayName: formatStudentDisplayName(u),
			IsActive:    u.IsActive,
			Role:        u.Role,
		})
	}
	h.writeJSON(w, http.StatusOK, map[string]interface{}{"students": out})
}

func (h *Handler) SubscriptionsAPI(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createSubscriptionAPI(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) SubscriptionByIDAPI(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/subscriptions/")
	idStr = strings.Trim(idStr, "/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid subscription id", http.StatusBadRequest)
		return
	}
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := h.requireCoach(w, r); !ok {
		return
	}
	if err := h.subscriptionService.DeleteSubscription(id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
}

func (h *Handler) createSubscriptionAPI(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireCoach(w, r); !ok {
		return
	}

	var req struct {
		StudentUserID int64  `json:"student_user_id"`
		Lessons       int    `json:"lessons"`
		EndDate       string `json:"end_date"` // YYYY-MM-DD required
		// Legacy field from early SPA: "1"|"12"|"16"
		Type string `json:"type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if req.StudentUserID <= 0 {
		http.Error(w, "student_user_id required", http.StatusBadRequest)
		return
	}

	lessons := req.Lessons
	if lessons == 0 && req.Type != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(req.Type)); err == nil {
			lessons = n
		}
	}
	switch lessons {
	case 1, 2, 3, 4, 8, 12, 16:
	default:
		http.Error(w, "lessons must be one of 1,2,3,4,8,12,16", http.StatusBadRequest)
		return
	}

	endDateStr := strings.TrimSpace(req.EndDate)
	if endDateStr == "" {
		http.Error(w, "end_date required (YYYY-MM-DD)", http.StatusBadRequest)
		return
	}
	endDate, err := time.ParseInLocation("2006-01-02", endDateStr, time.Local)
	if err != nil {
		http.Error(w, "invalid end_date, use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	student, err := h.studentService.GetStudentByUserID(req.StudentUserID)
	if err != nil || student == nil {
		http.Error(w, "Student not found", http.StatusBadRequest)
		return
	}

	if err := h.subscriptionService.CreateWithEndDate(student.ID, lessons, endDate); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
}
