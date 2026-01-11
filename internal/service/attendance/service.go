package attendance_service

import (
	"errors"
	"spectrum-club-bot/internal/models"
	"spectrum-club-bot/internal/repository"
	"spectrum-club-bot/internal/service"
	"time"
)

type attendanceService struct {
	attendanceRepo repository.AttendanceRepository
	scheduleRepo   repository.TrainingScheduleRepository
}

func NewAttendanceService(attendanceRepo repository.AttendanceRepository, scheduleRepo repository.TrainingScheduleRepository) service.AttendanceService {
	return &attendanceService{
		attendanceRepo: attendanceRepo,
		scheduleRepo:   scheduleRepo,
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
		return err
	}
	if attendance == nil {
		return errors.New("студент не записан на эту тренировку")
	}

	attendance.Attended = attended
	attendance.Notes = notes
	attendance.RecordedBy = &recordedBy
	attendance.RecordedAt = time.Now()

	return s.attendanceRepo.UpdateAttendance(attendance)
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
	// Нужно реализовать этот метод в репозитории
	return s.attendanceRepo.GetAttendanceByStudent(studentID, start, end)
}

func (s *attendanceService) GetStudentAttendanceForTraining(studentID, trainingID int) (*models.Attendance, error) {
	return s.attendanceRepo.GetStudentAttendanceForTraining(studentID, trainingID)
}
