package service

import (
	"spectrum-club-bot/internal/models"
	"time"
)

type UserService interface {
	RegisterOrUpdate(telegramID int64, firstName, lastName, username string, role string) (*models.User, error)
	GetUserProfile(telegramID int64) (*models.User, *models.Student, *models.Subscription, *models.Coach, error)
	SetRole(telegramID int64, role string) error
	RegisterAsCoach(telegramID int64, specialty, experience, description string) error

	GetAllStudents() ([]*models.User, error)
	GetByID(id int64) (*models.User, error)

	GetByTelegramID(telegramID int64) (*models.User, error)
}

type StudentService interface {
	GetStudentByUserID(userID int64) (*models.Student, error)
	UpdateAthleticTitle(studentID int64, athleticTitle string) error
	GetStudentWithUser(studentID int64) (*models.Student, *models.User, error)
}

type CoachService interface {
	RegisterCoach(userID int64, specialty, experience, description string) error
	GetCoachByUserID(userID int64) (*models.Coach, error)
	UpdateCoachProfile(coachID int64, specialty, experience, description string) error
	// GetAllCoaches() ([]*models.Coach, error)
	GetByCoachID(coachID int64) (*models.Coach, error)
}

type SubscriptionService interface {
	CreateSubscription(studentID int64, remainingLessons int, totalLessons int, durationDays int) error
	GetActiveSubscription(studentID int64) (*models.Subscription, error)
	UseLesson(subscriptionID int64) error
	ExtendSubscription(subscriptionID int64, additionalMonths int) error
	GetSubscriptionHistory(studentID int64) ([]*models.Subscription, error)

	////i did
	Create12Unlimited(studentID int64) error
	Create16For30Days(studentID int64) error
	Create1For30Days(studentID int64) error

	GetAll() ([]*models.Subscription, error)

	DeleteSubscription(subscriptionID int64) error
	GetSubscriptionsByStudentID(studentID int64) ([]*models.Subscription, error)
}

type TrainingGroupService interface {
	GetAllGroups() ([]models.TrainingGroup, error)
	GetGroupByID(id int) (*models.TrainingGroup, error)
	GetGroupsForAge(age int) ([]models.TrainingGroup, error)
}

// /////Создание тренировок и шаблонов для тренировок
type TrainingScheduleService interface {
	// Для тренеров
	CreateTraining(training *models.TrainingSchedule) error
	GetCoachSchedule(coachID int64, start, end time.Time) ([]models.TrainingSchedule, error)
	// Для студентов
	GetAvailableTrainings(studentID int, start, end time.Time) ([]models.TrainingSchedule, error)
	GetScheduleForDate(date time.Time) ([]models.TrainingSchedule, error)
	GetScheduleForWeek(startDate time.Time) ([]models.TrainingSchedule, error)
	//////для тренеров
	GetScheduleForGroup(groupID int, start, end time.Time) ([]models.TrainingSchedule, error)
	// GetTodaySchedule() ([]models.TrainingSchedule, error)
	// GetWeekSchedule() ([]models.TrainingSchedule, error)
	GetTrainingsByDateRange(start time.Time, end time.Time) ([]models.TrainingSchedule, error)
	UpdateTrainingPartial(id int, updates map[string]interface{}) error
	GetTrainingByID(id int) (*models.TrainingSchedule, error)
	DeleteTraining(id int) error

	WeekScheduleService
}

// ///шаблоны для тренеров................................
type WeekScheduleService interface {
	GetAllActiveTemplates() ([]models.WeekScheduleTemplate, error)
	GetTemplatesByGroup(groupID int) ([]models.WeekScheduleTemplate, error)
	CheckTrainingExists(groupID int, startTime time.Time) (bool, error)
	CreateTrainingsFromTemplates(weekStart time.Time, coachID int64, createdBy int64, weeksCount int) (int, error)
	GetTemplateByID(id int) (*models.WeekScheduleTemplate, error)
	UpdateTemplate(id int, updates map[string]interface{}) error
	DeactivateTemplate(id int) error
	ActivateTemplate(id int) error
	GetTemplatesForPreview() (string, error)
}

// ...............................
type AttendanceService interface {
	// Запись на тренировку
	SignUpForTraining(studentID, trainingID int) error
	// Отмена записи
	CancelSignUp(studentID, trainingID int) error
	// Для тренеров - отметка посещения
	MarkAttendance(trainingID, studentID, recordedBy int, attended bool, notes string) error
	// Просмотр записавшихся
	GetTrainingAttendees(trainingID int) ([]models.Attendance, error)
	// Статистика по тренировке
	GetTrainingStats(trainingID int) (present, absent, total int, err error)

	GetAttendanceByStudent(studentID int, start, end time.Time) ([]models.Attendance, error)

	GetStudentAttendanceForTraining(studentID, trainingID int) (*models.Attendance, error)

	GetParticipants(trainingID int) ([]models.AttendanceWithStudent, error)
	GetStudentSchedule(studentID int, start, end time.Time) ([]models.AttendanceWithTraining, error)
	CreateAttendance(attendance models.Attendance) error
	CancelAttendance(trainingID, studentID int) error
}
