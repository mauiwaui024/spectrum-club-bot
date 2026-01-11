package bot

import (
	"spectrum-club-bot/internal/models"
	"time"
)

type BotState int

const (
	StateDefault BotState = iota
	StateSelectingStudent
	StateSelectingSubscriptionType
	StateConfirming

	// Состояния для удаления абонемента
	StateSelectingStudentForDeletion
	StateSelectingSubscriptionForDeletion
	StateConfirmingSubscriptionDeletion

	// Состояния для добавления тренировки
	StateSelectingGroup
	StateSelectingDate
	StateSelectingTime
	StateSelectingDuration
	StateConfirmingTraining

	// Новые состояния для недельного расписания
	StateSelectingWeeksCount
	StateConfirmingWeeklySchedule

	StateSelectingTrainingDateToEdit /////
	StateSelectingTrainingToEdit
	StateSelectingFieldToEdit
	StateEditingTime
	StateEditingPlace
	StateEditingGroup // Новое состояние для редактирования группы
	StateEditingDate  // Новое состояние для редактирования даты
	StateConfirmingEdit

	StateSelectingScheduleDate   // Новое состояние для выбора даты расписания
	StateSelectingSchedulePeriod // Новое состояние для выбора периода расписания
	StateConfirmingDeletion      // Новое состояние для подтверждения удаления

	// Состояния для записи на тренировку
	StateSelectingTrainingDateToSignUp
	StateSelectingTrainingToSignUp
	StateConfirmingTrainingSignUp
)

type UserSession struct {
	State                    BotState
	SelectedStudentID        int64
	SelectedSubscriptionType string
	SelectedGroupID          int
	SelectedDate             time.Time
	SelectedStartTime        time.Time
	SelectedDuration         time.Duration
	TrainingDescription      string
	// Новые поля для удаления абонемента
	SelectedStudentForDeletion *models.User
	SelectedSubscriptionID     int64
	AvailableSubscriptions     []*models.Subscription
	// Новые поля для расписания
	ScheduleTrainings []models.TrainingSchedule
	ScheduleType      string // "date" или "period"
	ScheduleDate      time.Time
	ScheduleStartDate time.Time
	ScheduleEndDate   time.Time
	// Новое поле для недельного расписания
	WeeksCount int

	SelectedTrainingID     int
	AvailableTrainingsEdit []models.TrainingSchedule

	// Поля для записи на тренировку
	SelectedTrainingForSignUpID int
	AvailableTrainingsForSignUp []models.TrainingSchedule
	SelectedStudentForSignUpID  int

	StudentsForSelection []*models.User
}
