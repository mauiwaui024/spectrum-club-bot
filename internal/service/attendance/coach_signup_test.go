package attendance_service

import (
	"errors"
	"testing"
	"time"

	"spectrum-club-bot/internal/models"
	"spectrum-club-bot/internal/service"
)

func TestSignUpStudentByCoachIgnoresStartTime(t *testing.T) {
	coachID := int64(5)
	yesterday := time.Now().Add(-24 * time.Hour)
	att := &fakeAttendance{}
	sched := &fakeSchedule{training: &models.TrainingSchedule{
		ID:           9,
		CoachID:      &coachID,
		TrainingDate: yesterday,
		StartTime:    yesterday,
	}}
	subs := &fakeSubs{sub: &models.Subscription{ID: 1, StudentID: 20, RemainingLessons: 1}}
	svc := &attendanceService{
		attendanceRepo:      att,
		scheduleRepo:        sched,
		subscriptionService: subs,
		coachService:        &fakeCoach{coach: &models.Coach{ID: coachID, UserID: 1}},
		studentService:      &fakeStudent{student: &models.Student{ID: 20, UserID: 2}},
		userService:         &fakeUser{user: &models.User{ID: 2, Role: "student", IsActive: true}},
	}
	if err := svc.SignUpStudentByCoach(1, 20, 9); err != nil {
		t.Fatal(err)
	}
	if att.created == nil || att.created.Attended || att.created.Status != "registered" {
		t.Fatalf("attendance = %+v", att.created)
	}
}

func TestSignUpStudentByCoachRejectsOtherCoachAndMissingSubscription(t *testing.T) {
	coachID := int64(5)
	other := int64(8)
	svc := &attendanceService{
		attendanceRepo: &fakeAttendance{},
		scheduleRepo: &fakeSchedule{training: &models.TrainingSchedule{
			ID: 9, CoachID: &coachID, TrainingDate: time.Now(), StartTime: time.Now(),
		}},
		subscriptionService: &fakeSubs{},
		coachService:        &fakeCoach{coach: &models.Coach{ID: other, UserID: 1}},
		studentService:      &fakeStudent{student: &models.Student{ID: 20, UserID: 2}},
		userService:         &fakeUser{user: &models.User{ID: 2, Role: "student", IsActive: true}},
	}
	if err := svc.SignUpStudentByCoach(1, 20, 9); !errors.Is(err, service.ErrNotTrainingCoach) {
		t.Fatalf("got %v", err)
	}

	svc.coachService = &fakeCoach{coach: &models.Coach{ID: coachID, UserID: 1}}
	if err := svc.SignUpStudentByCoach(1, 20, 9); !errors.Is(err, service.ErrNoSubscription) {
		t.Fatalf("got %v", err)
	}
}

type fakeAttendance struct {
	created *models.Attendance
}

func (f *fakeAttendance) CreateAttendance(a *models.Attendance) error {
	f.created = a
	return nil
}
func (f *fakeAttendance) GetAttendanceByID(int) (*models.Attendance, error) { return nil, nil }
func (f *fakeAttendance) GetAttendanceByTraining(int) ([]models.Attendance, error) {
	return nil, nil
}
func (f *fakeAttendance) GetAttendanceByStudent(int, time.Time, time.Time) ([]models.Attendance, error) {
	return nil, nil
}
func (f *fakeAttendance) GetStudentAttendanceForTraining(int, int) (*models.Attendance, error) {
	return nil, nil
}
func (f *fakeAttendance) UpdateAttendance(*models.Attendance) error             { return nil }
func (f *fakeAttendance) DeleteAttendance(int) error                            { return nil }
func (f *fakeAttendance) GetTrainingAttendanceStats(int) (int, int, int, error) { return 0, 0, 0, nil }
func (f *fakeAttendance) CancelAttendance(int, int) error                       { return nil }
func (f *fakeAttendance) GetParticipants(int) ([]models.AttendanceWithStudent, error) {
	return nil, nil
}
func (f *fakeAttendance) GetStudentSchedule(int, time.Time, time.Time) ([]models.AttendanceWithTraining, error) {
	return nil, nil
}
func (f *fakeAttendance) CreateAttendanceRecord(models.Attendance) error { return nil }

type fakeSchedule struct {
	training *models.TrainingSchedule
}

func (f *fakeSchedule) CreateTraining(*models.TrainingSchedule) error { return nil }
func (f *fakeSchedule) GetTrainingByID(int) (*models.TrainingSchedule, error) {
	return f.training, nil
}
func (f *fakeSchedule) GetTrainingsByDate(time.Time) ([]models.TrainingSchedule, error) {
	return nil, nil
}
func (f *fakeSchedule) GetTrainingsByDateRange(time.Time, time.Time) ([]models.TrainingSchedule, error) {
	return nil, nil
}
func (f *fakeSchedule) GetTrainingsByGroup(int, time.Time, time.Time) ([]models.TrainingSchedule, error) {
	return nil, nil
}
func (f *fakeSchedule) GetTrainingsByCoach(int64, time.Time, time.Time) ([]models.TrainingSchedule, error) {
	return nil, nil
}
func (f *fakeSchedule) GetAvailableTrainingsForStudent(int, time.Time, time.Time) ([]models.TrainingSchedule, error) {
	return nil, nil
}
func (f *fakeSchedule) UpdateTraining(*models.TrainingSchedule) error { return nil }
func (f *fakeSchedule) UpdateTrainingPartial(int, map[string]interface{}) error {
	return nil
}
func (f *fakeSchedule) DeleteTraining(int) error { return nil }
func (f *fakeSchedule) IsCoachAvailable(int64, time.Time, time.Time, time.Time) (bool, error) {
	return true, nil
}
func (f *fakeSchedule) GetTrainingParticipantsCount(int) (int, error) { return 0, nil }
func (f *fakeSchedule) Exists(int, time.Time) (bool, error)           { return false, nil }
func (f *fakeSchedule) ExistsForCoach(int, int64, time.Time) (bool, error) {
	return false, nil
}

type fakeSubs struct {
	sub *models.Subscription
}

func (f *fakeSubs) CreateSubscription(int64, int, int, int) error { return nil }
func (f *fakeSubs) GetActiveSubscription(int64) (*models.Subscription, error) {
	return f.sub, nil
}
func (f *fakeSubs) UseLesson(int64) error               { return nil }
func (f *fakeSubs) ExtendSubscription(int64, int) error { return nil }
func (f *fakeSubs) GetSubscriptionHistory(int64) ([]*models.Subscription, error) {
	return nil, nil
}
func (f *fakeSubs) DecrementRemainingLessons(int64) error         { return nil }
func (f *fakeSubs) AddLessons(int64, int) error                   { return nil }
func (f *fakeSubs) RemoveLessons(int64, int) error                { return nil }
func (f *fakeSubs) SetEndDate(int64, time.Time) error             { return nil }
func (f *fakeSubs) Create12Unlimited(int64) error                 { return nil }
func (f *fakeSubs) Create16For30Days(int64) error                 { return nil }
func (f *fakeSubs) Create1For30Days(int64) error                  { return nil }
func (f *fakeSubs) CreateWithEndDate(int64, int, time.Time) error { return nil }
func (f *fakeSubs) GetAll() ([]*models.Subscription, error)       { return nil, nil }
func (f *fakeSubs) DeleteSubscription(int64) error                { return nil }
func (f *fakeSubs) GetSubscriptionsByStudentID(int64) ([]*models.Subscription, error) {
	return nil, nil
}

type fakeCoach struct{ coach *models.Coach }

func (f *fakeCoach) RegisterCoach(int64, string, string, string) error { return nil }
func (f *fakeCoach) GetCoachByUserID(int64) (*models.Coach, error)     { return f.coach, nil }
func (f *fakeCoach) UpdateCoachProfile(int64, *string, *string, *string) error {
	return nil
}
func (f *fakeCoach) GetByCoachID(int64) (*models.Coach, error) { return f.coach, nil }

type fakeStudent struct{ student *models.Student }

func (f *fakeStudent) GetStudentByUserID(int64) (*models.Student, error) { return f.student, nil }
func (f *fakeStudent) GetStudentByID(int64) (*models.Student, error)     { return f.student, nil }
func (f *fakeStudent) UpdateAthleticTitle(int64, *string) error          { return nil }
func (f *fakeStudent) UpdateCard(int64, *int64, *string, *string) error  { return nil }
func (f *fakeStudent) GetStudentWithUser(int64) (*models.Student, *models.User, error) {
	return f.student, nil, nil
}

type fakeUser struct{ user *models.User }

func (f *fakeUser) RegisterOrUpdate(int64, string, string, string, string) (*models.User, error) {
	return f.user, nil
}
func (f *fakeUser) CreateWebUser(string, string, string, string, string) (*models.User, error) {
	return f.user, nil
}
func (f *fakeUser) GetUserProfile(int64) (*models.User, *models.Student, *models.Subscription, *models.Coach, error) {
	return f.user, nil, nil, nil, nil
}
func (f *fakeUser) SetRole(int64, string) error                         { return nil }
func (f *fakeUser) RegisterAsCoach(int64, string, string, string) error { return nil }
func (f *fakeUser) UpdateProfile(int64, *string, *string) error         { return nil }
func (f *fakeUser) DeleteUser(int64) error                              { return nil }
func (f *fakeUser) GetAllStudents() ([]*models.User, error)             { return nil, nil }
func (f *fakeUser) GetByID(int64) (*models.User, error)                 { return f.user, nil }
func (f *fakeUser) GetByTelegramID(int64) (*models.User, error)         { return f.user, nil }
func (f *fakeUser) GetByUsername(string) (*models.User, error)          { return f.user, nil }
func (f *fakeUser) UsernameExists(string) (bool, error)                 { return false, nil }
func (f *fakeUser) SetBrowserPassword(int64, string) error              { return nil }
func (f *fakeUser) VerifyBrowserPassword(*models.User, string) bool     { return false }
func (f *fakeUser) SetActive(int64, bool) error                         { return nil }
func (f *fakeUser) SetRoleByUserID(int64, string) error                 { return nil }
func (f *fakeUser) EnsureCoachProfile(int64) error                      { return nil }
func (f *fakeUser) RecordPDConsent(int64, string) error                 { return nil }
func (f *fakeUser) ListUsersWithoutPassword() ([]*models.User, error) {
	return nil, nil
}
func (f *fakeUser) SetUsernameIfEmpty(int64, string) error { return nil }
