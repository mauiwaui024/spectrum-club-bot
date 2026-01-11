package subscription_service

import (
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
