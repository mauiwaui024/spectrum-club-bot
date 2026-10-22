package repository

import (
	"spectrum-club-bot/internal/models"
	"time"
)

type RefreshTokenRepository interface {
	Create(userID int64, tokenHash []byte, expiresAt time.Time) error
	GetByHash(tokenHash []byte) (*models.RefreshToken, error)
	DeleteByHash(tokenHash []byte) error
	DeleteAllForUser(userID int64) error
	DeleteExpired(before time.Time) error
}

type UserRepository interface {
	CreateOrUpdate(user *models.User) error
	CreateWithPassword(user *models.User, passwordHash string) error
	GetByTelegramID(telegramID int64) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	UsernameExists(username string) (bool, error)
	GetByID(id int64) (*models.User, error)
	SetPasswordHash(userID int64, passwordHash string) error
	UpdateRole(telegramID int64, role string) error
	UpdateRoleByUserID(userID int64, role string) error
	UpdateUser(userID int64, firstName, lastName *string) error
	SetActive(userID int64, active bool) error
	RecordPDConsent(userID int64, policyVersion string, at time.Time) error
	ListWithoutPassword() ([]*models.User, error)
	SetUsernameIfEmpty(userID int64, username string) error
	Delete(userID int64) error

	GetAllStudents() ([]*models.User, error)
}

type StudentRepository interface {
	Create(student *models.Student) error
	GetByUserID(userID int64) (*models.Student, error)
	GetByID(id int64) (*models.Student, error)
	Update(student *models.Student) error
	UpdateAthleticTitle(studentID int64, athleticTitle *string) error
	UpdateCard(studentID int64, groupID *int64, discipline *string, lastGrade *string) error
}

type CoachRepository interface {
	Create(coach *models.Coach) error
	GetByUserID(userID int64) (*models.Coach, error)
	GetByID(id int64) (*models.Coach, error)
	GetAll() ([]*models.Coach, error)
	Update(coach *models.Coach) error
	UpdateProfile(coachID int64, specialty, experience, description *string) error

	GetByCoachID(coachID int64) (*models.Coach, error)
}

type SubscriptionRepository interface {
	Create(subscription *models.Subscription) error
	// GetByStudentID(studentID int64) ([]*models.Subscription, error)
	GetByID(id int64) (*models.Subscription, error)
	GetActiveByStudentID(studentID int64) (*models.Subscription, error)
	GetHistoryByStudentID(studentID int64) ([]*models.Subscription, error)
	Update(subscription *models.Subscription) error
	DecrementRemainingLessons(studentID int64) error
	AddLessons(subscriptionID int64, count int) error
	RemoveLessons(subscriptionID int64, count int) error
	SetEndDate(subscriptionID int64, endDate time.Time) error

	///
	GetAll() ([]*models.Subscription, error)
	////
	GetByStudentID(studentID int64) ([]*models.Subscription, error)
	Delete(id int64) error
}

type TrainingGroupRepository interface {
	// Группы
	GetAllGroups() ([]models.TrainingGroup, error)
	GetGroupByID(id int) (*models.TrainingGroup, error)
	GetGroupByCode(code string) (*models.TrainingGroup, error)
	Create(group *models.TrainingGroup) error
	Update(id int, name, code, description string, ageMin int, ageMax *int) error
	Delete(id int) error
	CountFutureTrainings(groupID int) (int, error)
}

type TrainingScheduleRepository interface {
	// Тренировки
	CreateTraining(training *models.TrainingSchedule) error
	GetTrainingByID(id int) (*models.TrainingSchedule, error)
	GetTrainingsByDate(date time.Time) ([]models.TrainingSchedule, error)
	GetTrainingsByDateRange(start, end time.Time) ([]models.TrainingSchedule, error)
	GetTrainingsByGroup(groupID int, start, end time.Time) ([]models.TrainingSchedule, error)
	GetTrainingsByCoach(coachID int64, start, end time.Time) ([]models.TrainingSchedule, error)
	GetAvailableTrainingsForStudent(studentID int, start, end time.Time) ([]models.TrainingSchedule, error)
	UpdateTraining(training *models.TrainingSchedule) error
	UpdateTrainingPartial(id int, updates map[string]interface{}) error
	DeleteTraining(id int) error

	// Проверки
	IsCoachAvailable(coachID int64, date time.Time, startTime, endTime time.Time) (bool, error)
	GetTrainingParticipantsCount(trainingID int) (int, error)
	Exists(groupID int, startTime time.Time) (bool, error)
	ExistsForCoach(groupID int, coachID int64, startTime time.Time) (bool, error)
}

type WeekScheduleRepository interface {
	GetAllActive() ([]models.WeekScheduleTemplate, error)
	GetAll() ([]models.WeekScheduleTemplate, error)
	GetByGroupID(groupID int) ([]models.WeekScheduleTemplate, error)
	GetByID(id int) (*models.WeekScheduleTemplate, error)
	Create(template *models.WeekScheduleTemplate) error
	UpdatePartial(id int, updates map[string]interface{}) error
	Delete(id int) error
	Activate(id int) error
	Deactivate(id int) error
}

type AttendanceRepository interface {
	// Записи на тренировки
	CreateAttendance(attendance *models.Attendance) error
	GetAttendanceByID(id int) (*models.Attendance, error)
	GetAttendanceByTraining(trainingID int) ([]models.Attendance, error)
	GetAttendanceByStudent(studentID int, start, end time.Time) ([]models.Attendance, error)
	GetStudentAttendanceForTraining(studentID, trainingID int) (*models.Attendance, error)
	UpdateAttendance(attendance *models.Attendance) error
	DeleteAttendance(id int) error

	// Статистика
	GetTrainingAttendanceStats(trainingID int) (present, absent, total int, err error)

	//
	CancelAttendance(trainingID, studentID int) error
	GetParticipants(trainingID int) ([]models.AttendanceWithStudent, error)
	GetStudentSchedule(studentID int, start, end time.Time) ([]models.AttendanceWithTraining, error)
	CreateAttendanceRecord(attendance models.Attendance) error
}
