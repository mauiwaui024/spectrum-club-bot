package web

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func (h *Handler) AnalyticsSummaryAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := h.requireCoach(w, r); !ok {
		return
	}

	now := time.Now()
	from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	to := from.AddDate(0, 1, 0).Add(-time.Second)
	if v := r.URL.Query().Get("from"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			from = t
		}
	}
	if v := r.URL.Query().Get("to"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			to = t.Add(23*time.Hour + 59*time.Minute)
		}
	}

	trainings, err := h.scheduleService.GetTrainingsByDateRange(from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var bookings int
	var attended int
	var unmarked int

	for _, t := range trainings {
		parts, _ := h.attendanceService.GetParticipants(t.ID)
		bookings += len(parts)
		ended := !time.Now().Before(trainingEndTime(&t))
		for _, p := range parts {
			if p.Attended || p.Status == "attended" {
				attended++
				continue
			}
			if ended {
				unmarked++
			}
		}
	}

	students, _ := h.userService.GetAllStudents()
	activeSubs := 0
	remainingSum := 0
	expiring14 := 0
	cutoff := time.Now().AddDate(0, 0, 14)
	for _, u := range students {
		st, err := h.studentService.GetStudentByUserID(u.ID)
		if err != nil || st == nil {
			continue
		}
		sub, err := h.subscriptionService.GetActiveSubscription(st.ID)
		if err != nil || sub == nil {
			continue
		}
		activeSubs++
		remainingSum += sub.RemainingLessons
		if !sub.EndDate.After(cutoff) {
			expiring14++
		}
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"from":                       from.Format("2006-01-02"),
		"to":                         to.Format("2006-01-02"),
		"trainings_count":            len(trainings),
		"bookings_count":             bookings,
		"attended_count":             attended,
		"unmarked_count":             unmarked,
		"active_subscriptions":       activeSubs,
		"remaining_lessons_sum":      remainingSum,
		"expiring_subscriptions_14d": expiring14,
	})
}

func (h *Handler) AnalyticsAttendanceCSVAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := h.requireCoach(w, r); !ok {
		return
	}
	now := time.Now()
	year, month := now.Year(), int(now.Month())
	if y := r.URL.Query().Get("year"); y != "" {
		if n, err := strconv.Atoi(y); err == nil {
			year = n
		}
	}
	if m := r.URL.Query().Get("month"); m != "" {
		if n, err := strconv.Atoi(m); err == nil && n >= 1 && n <= 12 {
			month = n
		}
	}
	from := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	to := from.AddDate(0, 1, 0).Add(-time.Second)

	trainings, err := h.scheduleService.GetTrainingsByDateRange(from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=attendance_%d_%02d.csv", year, month))
	_, _ = w.Write([]byte{0xEF, 0xBB, 0xBF})
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"Дата", "Группа", "Ученик", "Статус", "Тренер"})
	for _, t := range trainings {
		parts, _ := h.attendanceService.GetParticipants(t.ID)
		for _, p := range parts {
			status := "Записан"
			if p.Attended || p.Status == "attended" {
				status = "Пришёл"
			}
			_ = cw.Write([]string{
				t.TrainingDate.Format("2006-01-02"),
				t.GroupName,
				p.StudentName,
				status,
				t.CoachName,
			})
		}
	}
	cw.Flush()
}

func (h *Handler) AnalyticsInactiveStudentsAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := h.requireCoach(w, r); !ok {
		return
	}
	weeks := 4
	if v := r.URL.Query().Get("weeks"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 52 {
			weeks = n
		}
	}
	since := time.Now().AddDate(0, 0, -7*weeks)

	students, err := h.userService.GetAllStudents()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	groups, _ := h.groupService.GetAllGroups()
	groupName := map[int64]string{}
	for _, g := range groups {
		groupName[int64(g.ID)] = g.Name
	}

	type row struct {
		UserID      int64  `json:"user_id"`
		StudentID   int64  `json:"student_id"`
		DisplayName string `json:"display_name"`
		GroupName   string `json:"group_name"`
	}
	out := make([]row, 0)
	for _, u := range students {
		if !u.IsActive {
			continue
		}
		st, err := h.studentService.GetStudentByUserID(u.ID)
		if err != nil || st == nil {
			continue
		}
		atts, err := h.attendanceService.GetAttendanceByStudent(int(st.ID), since, time.Now())
		if err != nil {
			continue
		}
		had := false
		for _, a := range atts {
			if a.Attended || a.Status == "attended" {
				had = true
				break
			}
		}
		if !had {
			name := ""
			if st.GroupID != nil {
				name = groupName[*st.GroupID]
			}
			out = append(out, row{
				UserID:      u.ID,
				StudentID:   st.ID,
				DisplayName: formatStudentDisplayName(u),
				GroupName:   name,
			})
		}
	}
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"weeks":    weeks,
		"since":    since.Format("2006-01-02"),
		"students": out,
	})
}
