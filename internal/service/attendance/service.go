package attendance_service

import (
	"errors"
	"fmt"
	"spectrum-club-bot/internal/models"
	"spectrum-club-bot/internal/repository"
	"spectrum-club-bot/internal/service"
	"time"
)

type attendanceService struct {
	attendanceRepo      repository.AttendanceRepository
	scheduleRepo        repository.TrainingScheduleRepository
	subscriptionService service.SubscriptionService
}

func NewAttendanceService(attendanceRepo repository.AttendanceRepository, scheduleRepo repository.TrainingScheduleRepository, subscriptionService service.SubscriptionService) service.AttendanceService {
	return &attendanceService{
		attendanceRepo:      attendanceRepo,
		scheduleRepo:        scheduleRepo,
		subscriptionService: subscriptionService,
	}
}

// Запись на тренировку
func (s *attendanceService) SignUpForTraining(studentID, trainingID int) error {
	// Проверяем, не записан ли уже
	existing, err := s.attendanceRepo.GetStudentAttendanceForTraining(studentID, trainingID)
	if err != nil {
		return err
	}
	if existing != nil {
		return errors.New("студент уже записан на эту тренировку")
	}

	// Проверяем доступность мест
	training, err := s.scheduleRepo.GetTrainingByID(trainingID)
	if err != nil {
		return err
	}

	if training.MaxParticipants != nil {
		currentCount, err := s.scheduleRepo.GetTrainingParticipantsCount(trainingID)
		if err != nil {
			return err
		}
		if currentCount >= *training.MaxParticipants {
			return errors.New("нет свободных мест на тренировку")
		}
	}

	attendance := &models.Attendance{
		TrainingID: trainingID,
		StudentID:  studentID,
		Attended:   false,
		RecordedAt: time.Now(),
	}

	return s.attendanceRepo.CreateAttendance(attendance)
}

// Отмена записи
func (s *attendanceService) CancelSignUp(studentID, trainingID int) error {
	attendance, err := s.attendanceRepo.GetStudentAttendanceForTraining(studentID, trainingID)
	if err != nil {
		return err
	}
	if attendance == nil {
		return errors.New("студент не записан на эту тренировку")
	}

	return s.attendanceRepo.DeleteAttendance(attendance.ID)
}

// Для тренеров - отметка посещения
func (s *attendanceService) MarkAttendance(trainingID, studentID, recordedBy int, attended bool, notes string) error {
	attendance, err := s.attendanceRepo.GetStudentAttendanceForTraining(studentID, trainingID)
	if err != nil {
		return fmt.Errorf("ошибка получения записи посещаемости: %w", err)
	}
	if attendance == nil {
		return errors.New("студент не записан на эту тренировку")
	}

	// Сохраняем старое значение для проверки необходимости списания абонемента
	oldAttended := attendance.Attended
	needsSubscriptionDeduction := attended && !oldAttended

	// Обновляем поля посещаемости
	attendance.Attended = attended
	attendance.Notes = notes
	attendance.RecordedBy = &recordedBy
	attendance.RecordedAt = time.Now()

	// Логирование для отладки
	fmt.Printf("[MarkAttendance] Обновление посещаемости: trainingID=%d, studentID=%d, attended=%v (было %v), attendance.ID=%d\n",
		trainingID, studentID, attended, oldAttended, attendance.ID)
	fmt.Printf("[MarkAttendance] attendance объект перед UpdateAttendance: ID=%d, Attended=%v, Notes=%s, RecordedBy=%v\n",
		attendance.ID, attendance.Attended, attendance.Notes, attendance.RecordedBy)

	// ВАЖНО: Сначала обновляем attendance в БД
	err = s.attendanceRepo.UpdateAttendance(attendance)
	if err != nil {
		fmt.Printf("[MarkAttendance] ОШИБКА при обновлении в БД: %v\n", err)
		return fmt.Errorf("ошибка обновления посещаемости в БД: %w", err)
	}

	fmt.Printf("[MarkAttendance] Посещаемость успешно обновлена для studentID=%d, attended=%v\n", studentID, attended)

	// Только после успешного обновления списываем абонемент
	if needsSubscriptionDeduction {
		err := s.subscriptionService.DecrementRemainingLessons(int64(studentID))
		if err != nil {
			// Логируем ошибку, но не откатываем обновление attendance
			// Это не критично для отметки посещаемости
			fmt.Printf("[MarkAttendance] Предупреждение: не удалось списать абонемент для studentID=%d: %v\n", studentID, err)
		} else {
			fmt.Printf("[MarkAttendance] Абонемент успешно списан для studentID=%d\n", studentID)
		}
	}

	return nil
}

// Просмотр записавшихся
func (s *attendanceService) GetTrainingAttendees(trainingID int) ([]models.Attendance, error) {
	return s.attendanceRepo.GetAttendanceByTraining(trainingID)
}

// Статистика по тренировке
func (s *attendanceService) GetTrainingStats(trainingID int) (present, absent, total int, err error) {
	return s.attendanceRepo.GetTrainingAttendanceStats(trainingID)
}

func (s *attendanceService) GetAttendanceByStudent(studentID int, start, end time.Time) ([]models.Attendance, error) {
	return s.attendanceRepo.GetAttendanceByStudent(studentID, start, end)
}

func (s *attendanceService) GetStudentAttendanceForTraining(studentID, trainingID int) (*models.Attendance, error) {
	return s.attendanceRepo.GetStudentAttendanceForTraining(studentID, trainingID)
}

func (s *attendanceService) GetParticipants(trainingID int) ([]models.AttendanceWithStudent, error) {
	return s.attendanceRepo.GetParticipants(trainingID)
}

func (s *attendanceService) GetStudentSchedule(studentID int, start, end time.Time) ([]models.AttendanceWithTraining, error) {
	return s.attendanceRepo.GetStudentSchedule(studentID, start, end)
}

func (s *attendanceService) CreateAttendance(attendance models.Attendance) error {
	// Проверяем, не записан ли уже студент
	existing, err := s.GetStudentAttendanceForTraining(attendance.StudentID, attendance.TrainingID)
	if err == nil && existing != nil && existing.Status == "registered" {
		return errors.New("student already registered for this training")
	}

	return s.attendanceRepo.CreateAttendanceRecord(attendance)
}

func (s *attendanceService) CancelAttendance(trainingID, studentID int) error {
	return s.attendanceRepo.CancelAttendance(trainingID, studentID)
}
