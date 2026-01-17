package student_service

import (
	"spectrum-club-bot/internal/models"
	"spectrum-club-bot/internal/repository"
	"spectrum-club-bot/internal/service"
)

type studentService struct {
	studentRepo repository.StudentRepository
	userRepo    repository.UserRepository
}

func NewStudentService(studentRepo repository.StudentRepository) service.StudentService {
	return &studentService{
		studentRepo: studentRepo,
	}
}

func (s *studentService) GetStudentByUserID(userID int64) (*models.Student, error) {
	return s.studentRepo.GetByUserID(userID)
}

func (s *studentService) GetStudentByID(studentID int64) (*models.Student, error) {
	return s.studentRepo.GetByID(studentID)
}

func (s *studentService) UpdateAthleticTitle(studentID int64, athleticTitle string) error {
	//TODO:IMPLEMENT
	student := &models.Student{
		ID:           studentID,
		AtleticTitle: athleticTitle,
	}
	return s.studentRepo.Update(student)
}

func (s *studentService) GetStudentWithUser(studentID int64) (*models.Student, *models.User, error) {
	// Этот метод будет полезен когда нам нужно получить студента вместе с данными пользователя
	// Пока заглушка - в реальной реализации нужно будет добавить соответствующий метод в репозиторий
	return nil, nil, nil
}
