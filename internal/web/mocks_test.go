package web

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"spectrum-club-bot/internal/models"
)

type mockUserService struct {
	usersByID       map[int64]*models.User
	usersByUsername map[string]*models.User
	created         []*models.User
	verifyOK        bool
}

func (m *mockUserService) RegisterOrUpdate(int64, string, string, string, string) (*models.User, error) {
	return nil, errors.New("not implemented")
}
func (m *mockUserService) CreateWebUser(username, password, firstName, lastName, role string) (*models.User, error) {
	u := &models.User{
		ID:        int64(100 + len(m.created)),
		Username:  username,
		FirstName: firstName,
		LastName:  lastName,
		Role:      role,
	}
	m.created = append(m.created, u)
	if m.usersByID == nil {
		m.usersByID = map[int64]*models.User{}
	}
	m.usersByID[u.ID] = u
	return u, nil
}
func (m *mockUserService) GetUserProfile(int64) (*models.User, *models.Student, *models.Subscription, *models.Coach, error) {
	return nil, nil, nil, nil, errors.New("not implemented")
}
func (m *mockUserService) SetRole(int64, string) error { return nil }
func (m *mockUserService) RegisterAsCoach(int64, string, string, string) error {
	return nil
}
func (m *mockUserService) UpdateProfile(int64, *string, *string) error { return nil }
func (m *mockUserService) GetAllStudents() ([]*models.User, error) {
	var out []*models.User
	for _, u := range m.usersByID {
		if u.Role == "student" {
			out = append(out, u)
		}
	}
	return out, nil
}
func (m *mockUserService) GetByID(id int64) (*models.User, error) {
	u, ok := m.usersByID[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return u, nil
}
func (m *mockUserService) GetByTelegramID(int64) (*models.User, error) {
	return nil, errors.New("not found")
}
func (m *mockUserService) GetByUsername(username string) (*models.User, error) {
	u, ok := m.usersByUsername[username]
	if !ok {
		return nil, errors.New("not found")
	}
	return u, nil
}
func (m *mockUserService) UsernameExists(string) (bool, error) { return false, nil }
func (m *mockUserService) SetBrowserPassword(userID int64, plainPassword string) error {
	u, ok := m.usersByID[userID]
	if !ok {
		return errors.New("not found")
	}
	hash := plainPassword
	u.PasswordHash = &hash
	return nil
}
func (m *mockUserService) VerifyBrowserPassword(*models.User, string) bool {
	return m.verifyOK
}
func (m *mockUserService) SetActive(int64, bool) error         { return nil }
func (m *mockUserService) SetRoleByUserID(int64, string) error { return nil }
func (m *mockUserService) EnsureCoachProfile(int64) error      { return nil }
func (m *mockUserService) RecordPDConsent(userID int64, policyVersion string) error {
	u, ok := m.usersByID[userID]
	if !ok {
		return errors.New("not found")
	}
	now := time.Now()
	u.PDConsentAt = &now
	u.PDPolicyVersion = &policyVersion
	return nil
}
func (m *mockUserService) ListUsersWithoutPassword() ([]*models.User, error) {
	var out []*models.User
	for _, u := range m.usersByID {
		if u.IsActive && (u.PasswordHash == nil || *u.PasswordHash == "") {
			out = append(out, u)
		}
	}
	return out, nil
}
func (m *mockUserService) SetUsernameIfEmpty(userID int64, username string) error {
	u, ok := m.usersByID[userID]
	if !ok {
		return errors.New("not found")
	}
	if strings.TrimSpace(u.Username) != "" {
		return errors.New("username already set")
	}
	u.Username = username
	return nil
}

type mockStudentService struct {
	byUserID map[int64]*models.Student
}

func (m *mockStudentService) GetStudentByUserID(userID int64) (*models.Student, error) {
	s, ok := m.byUserID[userID]
	if !ok {
		return nil, errors.New("not found")
	}
	return s, nil
}
func (m *mockStudentService) GetStudentByID(int64) (*models.Student, error) {
	return nil, errors.New("not found")
}
func (m *mockStudentService) UpdateAthleticTitle(int64, *string) error         { return nil }
func (m *mockStudentService) UpdateCard(int64, *int64, *string, *string) error { return nil }
func (m *mockStudentService) GetStudentWithUser(int64) (*models.Student, *models.User, error) {
	return nil, nil, errors.New("not found")
}

type mockCoachService struct {
	byUserID map[int64]*models.Coach
}

func (m *mockCoachService) RegisterCoach(int64, string, string, string) error { return nil }
func (m *mockCoachService) GetCoachByUserID(userID int64) (*models.Coach, error) {
	c, ok := m.byUserID[userID]
	if !ok {
		return nil, errors.New("not found")
	}
	return c, nil
}
func (m *mockCoachService) UpdateCoachProfile(int64, *string, *string, *string) error {
	return nil
}
func (m *mockCoachService) GetByCoachID(int64) (*models.Coach, error) {
	return nil, errors.New("not found")
}

type mockSubscriptionService struct {
	created []struct {
		studentID int64
		lessons   int
		endDate   time.Time
	}
	deleted []int64
}

func (m *mockSubscriptionService) CreateSubscription(int64, int, int, int) error {
	return nil
}
func (m *mockSubscriptionService) GetActiveSubscription(int64) (*models.Subscription, error) {
	return nil, nil
}
func (m *mockSubscriptionService) UseLesson(int64) error               { return nil }
func (m *mockSubscriptionService) ExtendSubscription(int64, int) error { return nil }
func (m *mockSubscriptionService) GetSubscriptionHistory(int64) ([]*models.Subscription, error) {
	return nil, nil
}
func (m *mockSubscriptionService) DecrementRemainingLessons(int64) error { return nil }
func (m *mockSubscriptionService) AddLessons(int64, int) error           { return nil }
func (m *mockSubscriptionService) RemoveLessons(int64, int) error        { return nil }
func (m *mockSubscriptionService) SetEndDate(int64, time.Time) error     { return nil }
func (m *mockSubscriptionService) Create12Unlimited(int64) error         { return nil }
func (m *mockSubscriptionService) Create16For30Days(int64) error         { return nil }
func (m *mockSubscriptionService) Create1For30Days(int64) error          { return nil }
func (m *mockSubscriptionService) CreateWithEndDate(studentID int64, lessons int, endDate time.Time) error {
	m.created = append(m.created, struct {
		studentID int64
		lessons   int
		endDate   time.Time
	}{studentID, lessons, endDate})
	return nil
}
func (m *mockSubscriptionService) GetAll() ([]*models.Subscription, error) { return nil, nil }
func (m *mockSubscriptionService) DeleteSubscription(id int64) error {
	m.deleted = append(m.deleted, id)
	return nil
}
func (m *mockSubscriptionService) GetSubscriptionsByStudentID(int64) ([]*models.Subscription, error) {
	return nil, nil
}

type mockGroupService struct {
	groups []models.TrainingGroup
}

func (m *mockGroupService) GetAllGroups() ([]models.TrainingGroup, error) { return m.groups, nil }
func (m *mockGroupService) GetGroupByID(id int) (*models.TrainingGroup, error) {
	for i := range m.groups {
		if m.groups[i].ID == id {
			return &m.groups[i], nil
		}
	}
	return nil, errors.New("not found")
}
func (m *mockGroupService) GetGroupsForAge(int) ([]models.TrainingGroup, error) {
	return nil, nil
}
func (m *mockGroupService) CreateGroup(g *models.TrainingGroup) error {
	g.ID = len(m.groups) + 1
	m.groups = append(m.groups, *g)
	return nil
}
func (m *mockGroupService) UpdateGroup(int, map[string]interface{}) error { return nil }
func (m *mockGroupService) DeleteGroup(int) error                         { return nil }

type mockScheduleService struct {
	created   []*models.TrainingSchedule
	trainings map[int]*models.TrainingSchedule
	deleted   []int
	fromTpl   int
}

func (m *mockScheduleService) CreateTraining(t *models.TrainingSchedule) error {
	t.ID = len(m.created) + 1
	cp := *t
	m.created = append(m.created, &cp)
	if m.trainings == nil {
		m.trainings = map[int]*models.TrainingSchedule{}
	}
	m.trainings[t.ID] = &cp
	return nil
}
func (m *mockScheduleService) GetCoachSchedule(int64, time.Time, time.Time) ([]models.TrainingSchedule, error) {
	var out []models.TrainingSchedule
	for _, t := range m.created {
		out = append(out, *t)
	}
	return out, nil
}
func (m *mockScheduleService) GetAvailableTrainings(int, time.Time, time.Time) ([]models.TrainingSchedule, error) {
	return nil, nil
}
func (m *mockScheduleService) GetScheduleForDate(time.Time) ([]models.TrainingSchedule, error) {
	return nil, nil
}
func (m *mockScheduleService) GetScheduleForWeek(time.Time) ([]models.TrainingSchedule, error) {
	return nil, nil
}
func (m *mockScheduleService) GetScheduleForGroup(int, time.Time, time.Time) ([]models.TrainingSchedule, error) {
	return nil, nil
}
func (m *mockScheduleService) GetTrainingsByDateRange(time.Time, time.Time) ([]models.TrainingSchedule, error) {
	return nil, nil
}
func (m *mockScheduleService) UpdateTrainingPartial(id int, updates map[string]interface{}) error {
	t := m.trainings[id]
	if t == nil {
		return errors.New("not found")
	}
	if d, ok := updates["description"].(string); ok {
		t.Description = d
	}
	return nil
}
func (m *mockScheduleService) GetTrainingByID(id int) (*models.TrainingSchedule, error) {
	t, ok := m.trainings[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return t, nil
}
func (m *mockScheduleService) DeleteTraining(id int) error {
	m.deleted = append(m.deleted, id)
	delete(m.trainings, id)
	return nil
}
func (m *mockScheduleService) GetAllActiveTemplates() ([]models.WeekScheduleTemplate, error) {
	return nil, nil
}
func (m *mockScheduleService) GetTemplatesByGroup(int) ([]models.WeekScheduleTemplate, error) {
	return nil, nil
}
func (m *mockScheduleService) CheckTrainingExists(int, time.Time) (bool, error) { return false, nil }
func (m *mockScheduleService) CreateTrainingsFromTemplates(time.Time, int64, int64, int) (int, error) {
	return m.fromTpl, nil
}
func (m *mockScheduleService) GetTemplateByID(int) (*models.WeekScheduleTemplate, error) {
	return nil, errors.New("not found")
}
func (m *mockScheduleService) GetAllTemplates() ([]models.WeekScheduleTemplate, error) {
	return nil, nil
}
func (m *mockScheduleService) CreateTemplate(*models.WeekScheduleTemplate) error { return nil }
func (m *mockScheduleService) DeleteTemplate(int) error                          { return nil }
func (m *mockScheduleService) UpdateTemplate(int, map[string]interface{}) error  { return nil }
func (m *mockScheduleService) DeactivateTemplate(int) error                      { return nil }
func (m *mockScheduleService) ActivateTemplate(int) error                        { return nil }
func (m *mockScheduleService) GetTemplatesForPreview() (string, error)           { return "", nil }

type mockAttendanceService struct {
	coachSignupErr error
}

func (m *mockAttendanceService) SignUpForTraining(int, int) error { return nil }
func (m *mockAttendanceService) SignUpStudentByCoach(int64, int64, int) error {
	return m.coachSignupErr
}
func (m *mockAttendanceService) CancelSignUp(int, int) error { return nil }
func (m *mockAttendanceService) MarkAttendance(int, int, int, bool, string) error {
	return nil
}
func (m *mockAttendanceService) GetTrainingAttendees(int) ([]models.Attendance, error) {
	return nil, nil
}
func (m *mockAttendanceService) GetTrainingStats(int) (int, int, int, error) {
	return 0, 0, 0, nil
}
func (m *mockAttendanceService) GetAttendanceByStudent(int, time.Time, time.Time) ([]models.Attendance, error) {
	return nil, nil
}
func (m *mockAttendanceService) GetStudentAttendanceForTraining(int, int) (*models.Attendance, error) {
	return nil, nil
}
func (m *mockAttendanceService) GetParticipants(int) ([]models.AttendanceWithStudent, error) {
	return nil, nil
}
func (m *mockAttendanceService) GetStudentSchedule(int, time.Time, time.Time) ([]models.AttendanceWithTraining, error) {
	return nil, nil
}
func (m *mockAttendanceService) CreateAttendance(models.Attendance) error { return nil }
func (m *mockAttendanceService) CancelAttendance(int, int) error          { return nil }
func (m *mockAttendanceService) UnmarkAttendance(int, int) error          { return nil }

type mockSessionService struct {
	access  map[string]int64
	refresh map[string]int64
	seq     int
}

func newMockSessionService() *mockSessionService {
	return &mockSessionService{
		access:  map[string]int64{},
		refresh: map[string]int64{},
	}
}

func (m *mockSessionService) IssueSession(userID int64) (string, string, error) {
	m.seq++
	a := fmt.Sprintf("access-%d-%d", userID, m.seq)
	r := fmt.Sprintf("refresh-%d-%d", userID, m.seq)
	m.access[a] = userID
	m.refresh[r] = userID
	return a, r, nil
}

func (m *mockSessionService) ResolveUserID(accessToken, refreshToken string) (int64, string, string, bool, error) {
	if userID, ok := m.access[accessToken]; ok {
		return userID, "", "", false, nil
	}
	if userID, ok := m.refresh[refreshToken]; ok {
		delete(m.refresh, refreshToken)
		na, nr, err := m.IssueSession(userID)
		if err != nil {
			return 0, "", "", false, err
		}
		return userID, na, nr, true, nil
	}
	return 0, "", "", false, errors.New("user not authenticated")
}

func (m *mockSessionService) RevokeRefresh(rawRefresh string) error {
	delete(m.refresh, rawRefresh)
	return nil
}

func (m *mockSessionService) RevokeAllForUser(userID int64) error {
	for k, v := range m.access {
		if v == userID {
			delete(m.access, k)
		}
	}
	for k, v := range m.refresh {
		if v == userID {
			delete(m.refresh, k)
		}
	}
	return nil
}

func testHandler(coachUserID int64) (*Handler, *mockUserService, *mockSubscriptionService, *mockScheduleService) {
	users := &mockUserService{
		usersByID: map[int64]*models.User{
			coachUserID: {ID: coachUserID, Role: "coach", Username: "coach1", FirstName: "Иван", LastName: "Петров", IsActive: true},
			2:           {ID: 2, Role: "student", Username: "stu1", FirstName: "Анна", LastName: "Яковлева", IsActive: true},
		},
	}
	students := &mockStudentService{
		byUserID: map[int64]*models.Student{
			2: {ID: 20, UserID: 2},
		},
	}
	coaches := &mockCoachService{
		byUserID: map[int64]*models.Coach{
			coachUserID: {ID: 5, UserID: coachUserID},
		},
	}
	subs := &mockSubscriptionService{}
	sched := &mockScheduleService{fromTpl: 3}
	groups := &mockGroupService{groups: []models.TrainingGroup{{ID: 1, Name: "НП"}}}
	sessions := newMockSessionService()

	h := &Handler{
		userService:         users,
		studentService:      students,
		coachService:        coaches,
		subscriptionService: subs,
		scheduleService:     sched,
		groupService:        groups,
		attendanceService:   &mockAttendanceService{},
		sessionService:      sessions,
		inviteSecret:        "test-invite-secret",
		baseURL:             "https://example.test",
		usedInviteJTIs:      map[string]time.Time{},
	}
	return h, users, subs, sched
}

func withCoachSession(h *Handler, coachUserID int64) (access, refresh string) {
	access, refresh, err := h.sessionService.IssueSession(coachUserID)
	if err != nil {
		panic(err)
	}
	return access, refresh
}

func addAuthCookies(req *http.Request, access, refresh string) {
	req.AddCookie(&http.Cookie{Name: accessCookieName, Value: access})
	req.AddCookie(&http.Cookie{Name: refreshCookieName, Value: refresh})
}
