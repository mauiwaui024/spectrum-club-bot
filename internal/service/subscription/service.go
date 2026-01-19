package subscription_service

import (
	"fmt"
	"spectrum-club-bot/internal/models"
	"spectrum-club-bot/internal/repository"
	"spectrum-club-bot/internal/service"
	"time"
)

type subscriptionService struct {
	subscriptionRepo repository.SubscriptionRepository
}

func NewSubscriptionService(subscriptionRepo repository.SubscriptionRepository) service.SubscriptionService {
	return &subscriptionService{
		subscriptionRepo: subscriptionRepo,
	}
}

func (s *subscriptionService) CreateSubscription(studentID int64, remainingLessons int, totalLessons int, durationDays int) error {
	subscription := &models.Subscription{
		StudentID:        studentID,
		StartDate:        time.Now(),
		EndDate:          time.Now().AddDate(0, 0, durationDays),
		TotalLessons:     totalLessons,
		RemainingLessons: remainingLessons,
		CreatedAt:        time.Now(),
	}
	return s.subscriptionRepo.Create(subscription)
}
func (s *subscriptionService) DeleteSubscription(subscriptionID int64) error {
	return s.subscriptionRepo.Delete(subscriptionID)
}

func (s *subscriptionService) GetSubscriptionsByStudentID(studentID int64) ([]*models.Subscription, error) {
	return s.subscriptionRepo.GetByStudentID(studentID)
}

func (s *subscriptionService) GetAll() ([]*models.Subscription, error) {
	return s.subscriptionRepo.GetAll()
}

func (s *subscriptionService) GetActiveSubscription(studentID int64) (*models.Subscription, error) {
	return s.subscriptionRepo.GetActiveByStudentID(studentID)
}

func (s *subscriptionService) UseLesson(subscriptionID int64) error {
	// subscription, err := s.subscriptionRepo.GetByStudentID(subscriptionID)
	// if err != nil {
	// 	return err
	// }

	// if subscription.UsedLessons < subscription.TotalLessons {
	// 	subscription.UsedLessons++

	// 	// Если использованы все занятия, деактивируем абонемент
	// 	if subscription.UsedLessons >= subscription.TotalLessons {
	// 		subscription.IsActive = false
	// 	}

	// 	return s.subscriptionRepo.Update(subscription)
	// }

	return nil
}

func (s *subscriptionService) ExtendSubscription(subscriptionID int64, additionalMonths int) error {
	// subscription, err := s.subscriptionRepo.GetByStudentID(subscriptionID)
	// if err != nil {
	// 	return err
	// }

	// // Продлеваем абонемент
	// subscription.EndDate = subscription.EndDate.AddDate(0, additionalMonths, 0)
	// subscription.IsActive = true

	// return s.subscriptionRepo.Update(subscription)
	return nil
}

func (s *subscriptionService) GetSubscriptionHistory(studentID int64) ([]*models.Subscription, error) {
	// Этот метод потребует добавления нового метода в репозиторий
	// Пока заглушка - вернем только активный абонемент
	activeSub, err := s.subscriptionRepo.GetActiveByStudentID(studentID)
	if err != nil {
		return nil, err
	}

	if activeSub != nil {
		return []*models.Subscription{activeSub}, nil
	}

	return []*models.Subscription{}, nil
}

func (s *subscriptionService) Create12Unlimited(studentID int64) error {
	subscription := &models.Subscription{
		StudentID:        studentID,
		StartDate:        time.Now(),
		EndDate:          time.Now().AddDate(2, 0, 0), // 2 года
		TotalLessons:     12,
		RemainingLessons: 12,
		CreatedAt:        time.Now(),
	}
	return s.subscriptionRepo.Create(subscription)
}

func (s *subscriptionService) Create16For30Days(studentID int64) error {
	subscription := &models.Subscription{
		StudentID:        studentID,
		StartDate:        time.Now(),
		EndDate:          time.Now().AddDate(0, 0, 30), // 30 дней
		TotalLessons:     16,
		RemainingLessons: 16,
		CreatedAt:        time.Now(),
	}
	return s.subscriptionRepo.Create(subscription)
}

func (s *subscriptionService) Create1For30Days(studentID int64) error {
	subscription := &models.Subscription{
		StudentID:        studentID,
		StartDate:        time.Now(),
		EndDate:          time.Now().AddDate(0, 0, 30), // 30 дней
		TotalLessons:     1,
		RemainingLessons: 1,
		CreatedAt:        time.Now(),
	}
	return s.subscriptionRepo.Create(subscription)
}

// DecrementRemainingLessons уменьшает remaining_lessons на 1 для активного абонемента ученика
func (s *subscriptionService) DecrementRemainingLessons(studentID int64) error {
	return s.subscriptionRepo.DecrementRemainingLessons(studentID)
}

// AddLessons добавляет занятия к абонементу
func (s *subscriptionService) AddLessons(subscriptionID int64, count int) error {
	subscription, err := s.subscriptionRepo.GetByID(subscriptionID)
	if err != nil {
		return err
	}
	if subscription == nil {
		return fmt.Errorf("абонемент с ID %d не найден", subscriptionID)
	}

	// Определяем максимальный лимит для типа абонемента
	// Типы: 1 (пробное), 12 (несгораемый), 16 (30 дней)
	var maxTypeLimit int
	if subscription.TotalLessons <= 1 {
		maxTypeLimit = 1
	} else if subscription.TotalLessons <= 12 {
		maxTypeLimit = 12
	} else {
		maxTypeLimit = 16
	}

	// Проверяем, что после добавления не превысим лимит типа
	newTotalLessons := subscription.TotalLessons + count
	if newTotalLessons > maxTypeLimit {
		possibleToAdd := maxTypeLimit - subscription.TotalLessons
		if possibleToAdd < 0 {
			possibleToAdd = 0
		}
		return fmt.Errorf("можно добавить максимум %d занятий (до исходного лимита абонемента %d занятий)", possibleToAdd, maxTypeLimit)
	}

	// Проверяем, что не добавляем больше, чем было использовано
	possibleToAdd := subscription.TotalLessons - subscription.RemainingLessons
	if count > possibleToAdd {
		return fmt.Errorf("можно добавить максимум %d занятий (до исходного лимита абонемента %d занятий)", possibleToAdd, maxTypeLimit)
	}

	return s.subscriptionRepo.AddLessons(subscriptionID, count)
}

// RemoveLessons снимает занятия с абонемента
func (s *subscriptionService) RemoveLessons(subscriptionID int64, count int) error {
	subscription, err := s.subscriptionRepo.GetByID(subscriptionID)
	if err != nil {
		return err
	}
	if subscription == nil {
		return fmt.Errorf("абонемент с ID %d не найден", subscriptionID)
	}

	// Проверяем, что достаточно занятий для снятия
	if count > subscription.RemainingLessons {
		return fmt.Errorf("недостаточно занятий для снятия. Доступно: %d", subscription.RemainingLessons)
	}

	return s.subscriptionRepo.RemoveLessons(subscriptionID, count)
}
