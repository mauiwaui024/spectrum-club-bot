package schedule_service

import (
	"errors"
	"fmt"
	"spectrum-club-bot/internal/models"
	"spectrum-club-bot/internal/repository"
	"spectrum-club-bot/internal/service"
	"strings"
	"time"
)

type trainingScheduleService struct {
	scheduleRepo     repository.TrainingScheduleRepository
	attendanceRepo   repository.AttendanceRepository
	weekScheduleRepo repository.WeekScheduleRepository
	groupRepo        repository.TrainingGroupRepository
}

func NewScheduleService(scheduleRepo repository.TrainingScheduleRepository, attendanceRepo repository.AttendanceRepository, weekScheduleRepo repository.WeekScheduleRepository, groupRepo repository.TrainingGroupRepository) service.TrainingScheduleService {
	return &trainingScheduleService{
		scheduleRepo:     scheduleRepo,
		attendanceRepo:   attendanceRepo,
		weekScheduleRepo: weekScheduleRepo,
		groupRepo:        groupRepo,
	}
}

// Для тренеров
func (s *trainingScheduleService) CreateTraining(training *models.TrainingSchedule) error {
	// Проверка доступности тренера
	if training.CoachID != nil {
		available, err := s.scheduleRepo.IsCoachAvailable(*training.CoachID, training.TrainingDate, training.StartTime, training.EndTime)
		if err != nil {
			return err
		}
		if !available {
			return errors.New("тренер уже занят в это время")
		}
	}

	return s.scheduleRepo.CreateTraining(training)
}

func (s *trainingScheduleService) DeleteTraining(id int) error {
	return s.scheduleRepo.DeleteTraining(id)
}

func (s *trainingScheduleService) GetCoachSchedule(coachID int64, start, end time.Time) ([]models.TrainingSchedule, error) {
	return s.scheduleRepo.GetTrainingsByCoach(coachID, start, end)
}

// /////////////////////////////////////////////

func (s *trainingScheduleService) GetScheduleForGroup(groupID int, start, end time.Time) ([]models.TrainingSchedule, error) {
	return s.scheduleRepo.GetTrainingsByGroup(groupID, start, end)
}

func (s *trainingScheduleService) GetTrainingsByDateRange(start time.Time, end time.Time) ([]models.TrainingSchedule, error) {
	return s.scheduleRepo.GetTrainingsByDateRange(start, end)
}

// ////////////////
// Для студентов
func (s *trainingScheduleService) GetAvailableTrainings(studentID int, start, end time.Time) ([]models.TrainingSchedule, error) {
	return s.scheduleRepo.GetAvailableTrainingsForStudent(studentID, start, end)
}

func (s *trainingScheduleService) GetScheduleForDate(date time.Time) ([]models.TrainingSchedule, error) {
	return s.scheduleRepo.GetTrainingsByDate(date)
}

func (s *trainingScheduleService) GetScheduleForWeek(startDate time.Time) ([]models.TrainingSchedule, error) {
	endDate := startDate.AddDate(0, 0, 7)
	return s.scheduleRepo.GetTrainingsByDateRange(startDate, endDate)
}

///////////////////////////////templates///////////////////////////////////

func (s *trainingScheduleService) GetAllActiveTemplates() ([]models.WeekScheduleTemplate, error) {
	return s.weekScheduleRepo.GetAllActive()
}

func (s *trainingScheduleService) GetTemplatesByGroup(groupID int) ([]models.WeekScheduleTemplate, error) {
	return s.weekScheduleRepo.GetByGroupID(groupID)
}

func (s *trainingScheduleService) CheckTrainingExists(groupID int, startTime time.Time) (bool, error) {
	////////////
	return s.scheduleRepo.Exists(groupID, startTime)
}

func (s *trainingScheduleService) CreateTrainingsFromTemplates(
	weekStart time.Time,
	coachID int64,
	createdBy int64,
	weeksCount int,
) (int, error) {

	// Получаем все активные шаблоны
	templates, err := s.weekScheduleRepo.GetAllActive()
	if err != nil {
		return 0, fmt.Errorf("ошибка получения шаблонов: %w", err)
	}

	if len(templates) == 0 {
		return 0, fmt.Errorf("нет активных шаблонов для создания тренировок")
	}

	createdCount := 0

	// Создаем тренировки для каждой недели
	for week := 0; week < weeksCount; week++ {
		weekDate := weekStart.AddDate(0, 0, week*7)

		for _, template := range templates {
			// Вычисляем дату тренировки (день недели)
			trainingDate := getDateForDayOfWeek(weekDate, template.DayOfWeek)

			startTimeStr := extractTimeOnly(template.StartTime)
			endTimeStr := extractTimeOnly(template.EndTime)

			// Теперь парсим как чистое время
			startTime, err := time.Parse("15:04:05", startTimeStr)
			if err != nil {
				fmt.Println("hhhhhhhhhhhhhhhh")
				fmt.Println(err)
				// Пропускаем шаблон с некорректным временем
				continue
			}

			// Парсим время окончания
			endTime, err := time.Parse("15:04:05", endTimeStr)
			if err != nil {
				fmt.Println("ffffffffffffffffffffffffffffffffff111")

				fmt.Println(err)
				continue
			}

			// Комбинируем дату и время
			trainingStart := time.Date(
				trainingDate.Year(),
				trainingDate.Month(),
				trainingDate.Day(),
				startTime.Hour(),
				startTime.Minute(),
				0, 0, time.Local,
			)

			trainingEnd := time.Date(
				trainingDate.Year(),
				trainingDate.Month(),
				trainingDate.Day(),
				endTime.Hour(),
				endTime.Minute(),
				0, 0, time.Local,
			)

			// Проверяем, не существует ли уже такая тренировка
			exists, err := s.scheduleRepo.Exists(template.GroupID, trainingStart)
			if err != nil {
				fmt.Println(err)
				fmt.Println("ffffffffffffffffffffffffffffffffff")
				continue
			}

			if exists {
				fmt.Println("12323213123124324325435345")
				continue
			}

			// Получаем информацию о группе
			group, err := s.groupRepo.GetGroupByID(template.GroupID)
			if err != nil {
				// Используем дефолтное описание если группа не найдена
				group = &models.TrainingGroup{Name: "Неизвестная группа"}
			}

			// Создаем тренировку
			training := &models.TrainingSchedule{
				GroupID:      template.GroupID,
				CoachID:      &coachID,
				TrainingDate: trainingStart,
				StartTime:    trainingStart,
				EndTime:      trainingEnd,
				Description:  fmt.Sprintf("%s - %s", template.Description, group.Name),
				CreatedBy:    &createdBy,
			}

			err = s.scheduleRepo.CreateTraining(training)
			if err != nil {
				fmt.Println(err)

				fmt.Println("45t4564564566566")
				continue
			}

			createdCount++
		}
	}

	return createdCount, nil
}

// GetTemplateByID возвращает шаблон по ID
func (s *trainingScheduleService) GetTemplateByID(id int) (*models.WeekScheduleTemplate, error) {
	return s.weekScheduleRepo.GetByID(id)
}

// UpdateTemplate обновляет шаблон (частичное обновление через COALESCE)
func (s *trainingScheduleService) UpdateTemplate(id int, updates map[string]interface{}) error {
	return s.weekScheduleRepo.UpdatePartial(id, updates)
}

// DeactivateTemplate деактивирует шаблон
func (s *trainingScheduleService) DeactivateTemplate(id int) error {
	return s.weekScheduleRepo.Deactivate(id)
}

// ActivateTemplate активирует шаблон
func (s *trainingScheduleService) ActivateTemplate(id int) error {
	return s.weekScheduleRepo.Activate(id)
}

// GetTemplatesForPreview возвращает шаблоны в удобном для просмотра формате
func (s *trainingScheduleService) GetTemplatesForPreview() (string, error) {
	templates, err := s.weekScheduleRepo.GetAllActive()
	if err != nil {
		return "", err
	}

	if len(templates) == 0 {
		return "📋 Нет активных шаблонов расписания", nil
	}

	// Группируем по дням недели
	dayTemplates := make(map[int][]models.WeekScheduleTemplate)
	for _, t := range templates {
		dayTemplates[t.DayOfWeek] = append(dayTemplates[t.DayOfWeek], t)
	}

	// Дни недели
	days := []string{"Понедельник", "Вторник", "Среда", "Четверг", "Пятница", "Суббота", "Воскресенье"}

	result := "📋 Активные шаблоны расписания:\n\n"

	// Проходим по всем дням недели
	for day := 1; day <= 7; day++ {
		if templates, exists := dayTemplates[day]; exists {
			result += fmt.Sprintf("📅 %s:\n", days[day-1])

			// Сортируем по времени
			// (можно добавить сортировку если нужно)
			for _, t := range templates {
				group, _ := s.groupRepo.GetGroupByID(t.GroupID)
				groupName := "Неизвестная группа"
				if group != nil {
					groupName = group.Name
				}

				result += fmt.Sprintf("  • %s-%s - %s (%s)\n",
					t.StartTime[:5],
					t.EndTime[:5],
					groupName,
					t.Description)
			}
			result += "\n"
		}
	}

	return result, nil
}

// Вспомогательная функция для получения даты по дню недели
func getDateForDayOfWeek(startOfWeek time.Time, dayOfWeek int) time.Time {
	// startOfWeek должен быть понедельником
	// dayOfWeek: 1=понедельник, 7=воскресенье

	// Если startOfWeek уже понедельник, добавляем (dayOfWeek-1) дней
	for startOfWeek.Weekday() != time.Monday {
		startOfWeek = startOfWeek.AddDate(0, 0, 1)
	}

	return startOfWeek.AddDate(0, 0, dayOfWeek-1)
}

func extractTimeOnly(timeStr string) string {
	// Пример: "0000-01-01T15:30:00Z" -> "15:30:00"

	// Находим 'T'
	idx := strings.Index(timeStr, "T")
	if idx == -1 {
		// Если 'T' нет, возвращаем как есть (уже должно быть время)
		return timeStr
	}

	// Берем часть после 'T' и удаляем 'Z' в конце если есть
	result := timeStr[idx+1:]
	if strings.HasSuffix(result, "Z") {
		result = result[:len(result)-1]
	}

	return result
}

func (s *trainingScheduleService) UpdateTrainingPartial(id int, updates map[string]interface{}) error {
	return s.scheduleRepo.UpdateTrainingPartial(id, updates)
}

func (s *trainingScheduleService) GetTrainingByID(id int) (*models.TrainingSchedule, error) {
	return s.scheduleRepo.GetTrainingByID(id)
}
