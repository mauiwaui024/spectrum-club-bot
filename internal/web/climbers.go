package web

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

type climberPatch struct {
	GroupID    *int64  `json:"group_id"`
	Discipline *string `json:"discipline"`
	LastGrade  *string `json:"last_grade"`
}

func (h *Handler) ClimbersAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	viewer, ok := h.requireUser(w, r)
	if !ok {
		return
	}
	users, err := h.userService.GetAllStudents()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	groups, err := h.groupService.GetAllGroups()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	groupName := map[int64]string{}
	for _, g := range groups {
		groupName[int64(g.ID)] = g.Name
	}

	cards := make([]map[string]interface{}, 0, len(users))
	for _, u := range users {
		if u == nil || !u.IsActive {
			continue
		}
		st, err := h.studentService.GetStudentByUserID(u.ID)
		if err != nil || st == nil {
			continue
		}
		card := map[string]interface{}{
			"student_id": st.ID,
			"user_id":    u.ID,
			"name":       formatStudentDisplayName(u),
			"group_id":   st.GroupID,
			"group_name": "",
			"discipline": "",
			"last_grade": "",
			"is_self":    u.ID == viewer.ID,
		}
		if st.GroupID != nil {
			card["group_name"] = groupName[*st.GroupID]
		}
		if st.Discipline != nil {
			card["discipline"] = *st.Discipline
		}
		if st.LastGrade != nil {
			card["last_grade"] = *st.LastGrade
		}
		if u.ID == viewer.ID {
			if sub, err := h.subscriptionService.GetActiveSubscription(st.ID); err == nil && sub != nil {
				card["remaining_lessons"] = sub.RemainingLessons
			} else {
				card["remaining_lessons"] = nil
			}
		}
		cards = append(cards, card)
	}
	sort.Slice(cards, func(i, j int) bool {
		return strings.ToLower(cards[i]["name"].(string)) < strings.ToLower(cards[j]["name"].(string))
	})
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"climbers": cards,
		"is_coach": viewer.Role == "coach",
	})
}

func (h *Handler) ClimberByIDAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := h.requireCoach(w, r); !ok {
		return
	}
	id, err := strconv.ParseInt(strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/climbers/"), "/"), 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid student id", http.StatusBadRequest)
		return
	}
	st, err := h.studentService.GetStudentByID(id)
	if err != nil || st == nil {
		http.Error(w, "Student not found", http.StatusNotFound)
		return
	}
	var body climberPatch
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	var discipline *string
	if body.Discipline != nil {
		value := strings.TrimSpace(*body.Discipline)
		if value != "" && value != "boulder" && value != "lead" && value != "both" {
			http.Error(w, "Неизвестная дисциплина", http.StatusBadRequest)
			return
		}
		if value == "" {
			discipline = nil
		} else {
			discipline = &value
		}
	}
	var grade *string
	if body.LastGrade != nil {
		value := strings.TrimSpace(*body.LastGrade)
		if len([]rune(value)) > 64 {
			http.Error(w, "Слишком длинная категория", http.StatusBadRequest)
			return
		}
		if value == "" {
			grade = nil
		} else {
			grade = &value
		}
	}
	var groupID *int64
	if body.GroupID != nil && *body.GroupID > 0 {
		g, err := h.groupService.GetGroupByID(int(*body.GroupID))
		if err != nil || g == nil {
			http.Error(w, "Группа не найдена", http.StatusBadRequest)
			return
		}
		groupID = body.GroupID
	}
	if err := h.studentService.UpdateCard(id, groupID, discipline, grade); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}
