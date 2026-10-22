package user_service

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"spectrum-club-bot/internal/models"
	"spectrum-club-bot/internal/repository"
	"spectrum-club-bot/internal/service"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	userRepo         repository.UserRepository
	studentRepo      repository.StudentRepository
	coachRepo        repository.CoachRepository
	subscriptionRepo repository.SubscriptionRepository
}

func NewUserService(
	userRepo repository.UserRepository,
	studentRepo repository.StudentRepository,
	coachRepo repository.CoachRepository,
	subscriptionRepo repository.SubscriptionRepository,
) service.UserService {
	return &userService{
		userRepo:         userRepo,
		studentRepo:      studentRepo,
		coachRepo:        coachRepo,
		subscriptionRepo: subscriptionRepo,
	}
}

func (s *userService) UsernameExists(username string) (bool, error) {
	return s.userRepo.UsernameExists(strings.TrimSpace(username))
}

// CreateWebUser создаёт пользователя без Telegram: dummy telegram_id + password_hash.
func (s *userService) CreateWebUser(username, password, firstName, lastName, role string) (*models.User, error) {
	username = strings.TrimSpace(username)
	firstName = strings.TrimSpace(firstName)
	lastName = strings.TrimSpace(lastName)
	role = strings.TrimSpace(role)

	if username == "" {
		return nil, fmt.Errorf("username is required")
	}
	if role != "student" && role != "coach" {
		return nil, fmt.Errorf("role must be student or coach")
	}
	if len(strings.TrimSpace(password)) < 8 {
		return nil, fmt.Errorf("password must be at least 8 characters")
	}

	exists, err := s.userRepo.UsernameExists(username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("username already taken")
	}

	passwordHashBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// UNIQUE NOT NULL telegram_id: отрицательный синтетический id вне диапазона Telegram.
	dummyTelegramID := -time.Now().UnixNano()
	if dummyTelegramID >= 0 {
		dummyTelegramID = -dummyTelegramID - 1
	}

	user := &models.User{
		TelegramID:   dummyTelegramID,
		FirstName:    firstName,
		LastName:     lastName,
		Username:     username,
		Role:         role,
		IsActive:     true,
		RegisteredAt: time.Now(),
	}

	if err := s.userRepo.CreateWithPassword(user, string(passwordHashBytes)); err != nil {
		return nil, err
	}

	student := &models.Student{
		UserID:    user.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.studentRepo.Create(student); err != nil {
		return nil, err
	}

	if role == "coach" {
		if err := s.userRepo.UpdateRoleByUserID(user.ID, "coach"); err != nil {
			return nil, err
		}
		user.Role = "coach"
		coach := &models.Coach{
			UserID:      user.ID,
			Specialty:   "Скалолазание",
			Experience:  "",
			Description: "",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if err := s.coachRepo.Create(coach); err != nil {
			return nil, err
		}
	}

	return user, nil
}

func (s *userService) RegisterOrUpdate(telegramID int64, firstName, lastName, username string, role string) (*models.User, error) {
	user := &models.User{
		TelegramID:   telegramID,
		FirstName:    firstName,
		LastName:     lastName,
		Username:     username,
		Role:         role,
		RegisteredAt: time.Now(),
	}

	err := s.userRepo.CreateOrUpdate(user)
	if err != nil {
		return nil, err
	}

	existingStudent, err := s.studentRepo.GetByUserID(user.ID)
	if err != nil || existingStudent == nil {
		// Создаем нового студента
		student := &models.Student{
			UserID:    user.ID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err = s.studentRepo.Create(student)
		if err != nil {
			return nil, err
		}

	}

	return user, nil
}

func (s *userService) GetUserProfile(telegramID int64) (*models.User, *models.Student, *models.Subscription, *models.Coach, error) {
	user, err := s.userRepo.GetByTelegramID(telegramID)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	var student *models.Student
	var subscription *models.Subscription
	var coach *models.Coach

	if user.Role == "student" {
		student, err = s.studentRepo.GetByUserID(user.ID)
		if err != nil {
			return user, nil, nil, nil, nil
		}

		if student != nil {
			subscription, err = s.subscriptionRepo.GetActiveByStudentID(student.ID)
			if err != nil {
				return user, student, nil, nil, nil
			}
		}
	} else if user.Role == "coach" {
		coach, err = s.coachRepo.GetByUserID(user.ID)
		if err != nil {
			return user, nil, nil, nil, nil
		}
	}

	return user, student, subscription, coach, nil
}

func (s *userService) SetRole(telegramID int64, role string) error {
	return s.userRepo.UpdateRole(telegramID, role)
}

func (s *userService) RegisterAsCoach(telegramID int64, specialty, experience, description string) error {
	user, err := s.userRepo.GetByTelegramID(telegramID)
	if err != nil {
		return err
	}

	// Обновляем роль пользователя
	err = s.userRepo.UpdateRole(telegramID, "coach")
	if err != nil {
		return err
	}

	// Создаем запись тренера
	coach := &models.Coach{
		UserID:      user.ID,
		Specialty:   specialty,
		Experience:  experience,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	return s.coachRepo.Create(coach)
}

// В user_service.go
func (s *userService) GetAllStudents() ([]*models.User, error) {
	return s.userRepo.GetAllStudents()
}

func (s *userService) GetByID(id int64) (*models.User, error) {
	return s.userRepo.GetByID(id)
}

func (s *userService) GetByTelegramID(telegramID int64) (*models.User, error) {
	return s.userRepo.GetByTelegramID(telegramID)
}

func (s *userService) GetByUsername(username string) (*models.User, error) {
	return s.userRepo.GetByUsername(username)
}

func (s *userService) UpdateProfile(userID int64, firstName, lastName *string) error {
	return s.userRepo.UpdateUser(userID, firstName, lastName)
}

func (s *userService) DeleteUser(userID int64) error {
	return s.userRepo.Delete(userID)
}

func (s *userService) SetBrowserPassword(userID int64, plainPassword string) error {
	plainPassword = strings.TrimSpace(plainPassword)
	if len(plainPassword) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	passwordHashBytes, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.userRepo.SetPasswordHash(userID, string(passwordHashBytes))
}

func (s *userService) SetActive(userID int64, active bool) error {
	return s.userRepo.SetActive(userID, active)
}

func (s *userService) SetRoleByUserID(userID int64, role string) error {
	role = strings.TrimSpace(role)
	if role != "student" && role != "coach" {
		return fmt.Errorf("role must be student or coach")
	}
	return s.userRepo.UpdateRoleByUserID(userID, role)
}

func (s *userService) ListUsersWithoutPassword() ([]*models.User, error) {
	return s.userRepo.ListWithoutPassword()
}

func (s *userService) SetUsernameIfEmpty(userID int64, username string) error {
	username = strings.TrimSpace(username)
	if username == "" {
		return fmt.Errorf("укажите логин")
	}
	exists, err := s.userRepo.UsernameExists(username)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("username already taken")
	}
	return s.userRepo.SetUsernameIfEmpty(userID, username)
}

func (s *userService) RecordPDConsent(userID int64, policyVersion string) error {
	if policyVersion != models.PDPolicyVersion {
		return fmt.Errorf("неизвестная версия политики")
	}
	return s.userRepo.RecordPDConsent(userID, policyVersion, time.Now())
}

func (s *userService) EnsureCoachProfile(userID int64) error {
	existing, err := s.coachRepo.GetByUserID(userID)
	if err == nil && existing != nil {
		return nil
	}
	coach := &models.Coach{
		UserID:      userID,
		Specialty:   "Скалолазание",
		Experience:  "",
		Description: "",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	return s.coachRepo.Create(coach)
}

func (s *userService) VerifyBrowserPassword(user *models.User, plainPassword string) bool {
	if user == nil || user.PasswordHash == nil || *user.PasswordHash == "" {
		return false
	}

	// Primary format: bcrypt hash.
	if strings.HasPrefix(*user.PasswordHash, "$2a$") || strings.HasPrefix(*user.PasswordHash, "$2b$") || strings.HasPrefix(*user.PasswordHash, "$2y$") {
		return bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(plainPassword)) == nil
	}

	// Backward compatibility for already-issued legacy hashes.
	parts := strings.Split(*user.PasswordHash, "$")
	if len(parts) != 3 || parts[0] != "sha256" {
		return false
	}
	sum := sha256.Sum256([]byte(parts[1] + ":" + plainPassword))
	actualHashHex := hex.EncodeToString(sum[:])
	expectedHashHex := parts[2]

	return subtle.ConstantTimeCompare([]byte(actualHashHex), []byte(expectedHashHex)) == 1
}
