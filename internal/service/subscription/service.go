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
	return s.CreateWithEndDate(studentID, 12, time.Now().AddDate(2, 0, 0))
}

func (s *subscriptionService) Create16For30Days(studentID int64) error {
	return s.CreateWithEndDate(studentID, 16, time.Now().AddDate(0, 0, 30))
}

func (s *subscriptionService) Create1For30Days(studentID int64) error {
	return s.CreateWithEndDate(studentID, 1, time.Now().AddDate(0, 0, 30))
}

func (s *subscriptionService) CreateWithEndDate(studentID int64, lessons int, endDate time.Time) error {
	switch lessons {
	case 1, 2, 3, 4, 8, 12, 16:
	default:
		return fmt.Errorf("неподдерживаемое число занятий: %d", lessons)
	}
	now := time.Now()
	endDay := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, endDate.Location())
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if endDay.Before(today) {
		return fmt.Errorf("дата окончания не может быть в прошлом")
	}
	subscription := &models.Subscription{
		StudentID:        studentID,
		StartDate:        now,
		EndDate:          endDay,
		TotalLessons:     lessons,
		RemainingLessons: lessons,
		CreatedAt:        now,
	}
	return s.subscriptionRepo.Create(subscription)
}

// MaxLessonsForSubscriptionType возвращает потолок занятий для абонемента по текущему total_lessons.
func MaxLessonsForSubscriptionType(totalLessons int) int {
	switch totalLessons {
	case 1, 2, 3, 4, 8, 12, 16:
		return totalLessons
	default:
		// Ближайший допустимый потолок сверху из каталога.
		for _, limit := range []int{1, 2, 3, 4, 8, 12, 16} {
			if totalLessons <= limit {
				return limit
			}
		}
		return 16
	}
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

	maxTypeLimit := MaxLessonsForSubscriptionType(subscription.TotalLessons)

	// Максимум, который можно добавить до исходного количества (использованные занятия)
	possibleToAddUsed := subscription.TotalLessons - subscription.RemainingLessons

	// Если total_lessons превышает лимит типа, можно добавить только до лимита типа
	if subscription.TotalLessons > maxTypeLimit {
		// Максимум до лимита типа: сколько можно добавить, чтобы total_lessons не превысил maxTypeLimit
		// Но также учитываем, что можно вернуть использованные занятия до исходного лимита
		// Если total_lessons уже превышает лимит, можно вернуть использованные занятия,
		// но так, чтобы total_lessons не превысил maxTypeLimit
		// Максимум, который можно добавить до лимита типа по total_lessons
		maxTotalAtLimit := maxTypeLimit
		currentTotal := subscription.TotalLessons
		possibleToAddToTotalLimit := maxTotalAtLimit - currentTotal

		// Максимум до лимита типа по remaining_lessons
		maxRemainingAtLimit := maxTypeLimit
		currentRemaining := subscription.RemainingLessons
		possibleToAddToRemainingLimit := maxRemainingAtLimit - currentRemaining

		// Берем минимум из всех ограничений
		possibleToAdd := possibleToAddUsed
		if possibleToAddToTotalLimit < possibleToAdd {
			possibleToAdd = possibleToAddToTotalLimit
		}
		if possibleToAddToRemainingLimit < possibleToAdd && possibleToAddToRemainingLimit >= 0 {
			possibleToAdd = possibleToAddToRemainingLimit
		}
		if possibleToAdd < 0 {
			possibleToAdd = 0
		}
		if count > possibleToAdd {
			return fmt.Errorf("можно добавить максимум %d занятий (до исходного лимита абонемента %d занятий)", possibleToAdd, maxTypeLimit)
		}
	} else if subscription.TotalLessons == maxTypeLimit {
		// Точно на лимите типа, можно только вернуть использованные занятия
		if count > possibleToAddUsed {
			return fmt.Errorf("можно добавить максимум %d занятий (до исходного лимита абонемента %d занятий)", possibleToAddUsed, maxTypeLimit)
		}
	} else {
		// Ниже лимита типа, проверяем оба ограничения
		// Максимум до лимита типа
		possibleToAddToLimit := maxTypeLimit - subscription.TotalLessons
		// Берем минимум из двух ограничений
		possibleToAdd := possibleToAddUsed
		if possibleToAddToLimit < possibleToAddUsed {
			possibleToAdd = possibleToAddToLimit
		}
		if count > possibleToAdd {
			return fmt.Errorf("можно добавить максимум %d занятий (до исходного лимита абонемента %d занятий)", possibleToAdd, maxTypeLimit)
		}
	}

	return s.subscriptionRepo.AddLessons(subscriptionID, count)
}

func (s *subscriptionService) SetEndDate(subscriptionID int64, endDate time.Time) error {
	subscription, err := s.subscriptionRepo.GetByID(subscriptionID)
	if err != nil {
		return err
	}
	if subscription == nil {
		return fmt.Errorf("абонемент с ID %d не найден", subscriptionID)
	}
	now := time.Now()
	endDay := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, endDate.Location())
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if endDay.Before(today) {
		return fmt.Errorf("дата окончания не может быть в прошлом")
	}
	return s.subscriptionRepo.SetEndDate(subscriptionID, endDay)
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

	// Разрешаем ручное снятие до порога -2 включительно.
	maxRemovable := subscription.RemainingLessons + 2
	if maxRemovable < 0 {
		maxRemovable = 0
	}
	if count > maxRemovable {
		return fmt.Errorf("недостаточно занятий для снятия. Можно снять максимум: %d", maxRemovable)
	}

	return s.subscriptionRepo.RemoveLessons(subscriptionID, count)
}
