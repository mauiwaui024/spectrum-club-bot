package repository

import (
	"spectrum-club-bot/internal/models"
	"time"
)

type UserRepository interface {
	CreateOrUpdate(user *models.User) error
	GetByTelegramID(telegramID int64) (*models.User, error)
	GetByID(id int64) (*models.User, error)
	UpdateRole(telegramID int64, role string) error

	GetAllStudents() ([]*models.User, error)
}

type StudentRepository interface {
	Create(student *models.Student) error
	GetByUserID(userID int64) (*models.Student, error)
	GetByID(id int64) (*models.Student, error)
	Update(student *models.Student) error
}

type CoachRepository interface {
	Create(coach *models.Coach) error
	GetByUserID(userID int64) (*models.Coach, error)
	GetByID(id int64) (*models.Coach, error)
	GetAll() ([]*models.Coach, error)
	Update(coach *models.Coach) error

	GetByCoachID(coachID int64) (*models.Coach, error)
}

type SubscriptionRepository interface {
	Create(subscription *models.Subscription) error
	// GetByStudentID(studentID int64) ([]*models.Subscription, error)
	GetByID(id int64) (*models.Subscription, error)
	GetActiveByStudentID(studentID int64) (*models.Subscription, error)
	GetHistoryByStudentID(studentID int64) ([]*models.Subscription, error)
	Update(subscription *models.Subscription) error

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
}

type WeekScheduleRepository interface {
	GetAllActive() ([]models.WeekScheduleTemplate, error)
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
}
