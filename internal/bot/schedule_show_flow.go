package bot

import (
	"fmt"
	"sort"
	"spectrum-club-bot/internal/models"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) handleMySchedule(chatID int64, user *models.User) {
	if user.Role != "coach" {
		b.sendError(chatID, "❌ Эта функция доступна только тренерам")
		return
	}

	// Показываем меню выбора типа расписания
	msg := tgbotapi.NewMessage(chatID, "📅 *Выберите тип расписания:*\n\nНа какую дату или период вы хотите посмотреть расписание?")
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📅 На конкретную дату"),
			tgbotapi.NewKeyboardButton("📆 На период"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("◀️ Назад"),
		),
	)

	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}

// ///////////////////////////////////////eeeeeeeeeeeeeeeeeeeee
func (b *Bot) handleScheduleTypeSelection(chatID int64, messageText string) {
	session := b.getOrCreateSession(chatID)

	switch messageText {
	case "📅 На конкретную дату":
		session.State = StateSelectingScheduleDate
		session.ScheduleType = "date"
		b.showDateInputForSchedule(chatID, "📅 *Введите дату для просмотра расписания:*\n\nФормат: ДД.ММ.ГГГГ\nПример: 15.12.2024")

	case "📆 На период":
		session.State = StateSelectingSchedulePeriod
		session.ScheduleType = "period"
		b.showPeriodInputForSchedule(chatID)

	case "◀️ Назад":
		b.showScheduleManagementMenu(chatID, nil)

	default:
		b.sendError(chatID, "❌ Пожалуйста, выберите один из вариантов")
	}
}
func (b *Bot) handleScheduleDateInput(chatID int64, messageText string) {
	session := b.getOrCreateSession(chatID)
	if session.State != StateSelectingScheduleDate {
		return
	}

	if messageText == "❌ Отмена" {
		b.showScheduleManagementMenu(chatID, nil)
		b.resetSession(chatID)
		return
	}

	var selectedDate time.Time
	now := time.Now()

	switch messageText {
	case "Сегодня":
		selectedDate = now
	case "Завтра":
		selectedDate = now.AddDate(0, 0, 1)
	case "Послезавтра":
		selectedDate = now.AddDate(0, 0, 2)
	case "Через неделю":
		selectedDate = now.AddDate(0, 0, 7)
	default:
		parsedDate, err := time.Parse("02.01.2006", messageText)
		if err != nil {
			b.sendError(chatID, "❌ Неверный формат даты. Используйте ДД.ММ.ГГГГ")
			return
		}
		selectedDate = parsedDate
	}

	// Сохраняем дату
	session.ScheduleDate = selectedDate

	// Получаем тренировки на эту дату
	b.showScheduleForDate(chatID, selectedDate)
}

func (b *Bot) handleSchedulePeriodInput(chatID int64, messageText string) {
	session := b.getOrCreateSession(chatID)
	if session.State != StateSelectingSchedulePeriod {
		return
	}

	if messageText == "❌ Отмена" {
		b.showScheduleManagementMenu(chatID, nil)
		b.resetSession(chatID)
		return
	}

	now := time.Now()
	var startDate, endDate time.Time

	switch messageText {
	case "Эта неделя":
		// Начало текущей недели (понедельник)
		weekday := int(now.Weekday())
		if weekday == 0 { // Воскресенье
			weekday = 7
		}
		startDate = now.AddDate(0, 0, -(weekday - 1))
		endDate = startDate.AddDate(0, 0, 6)

	case "Следующая неделя":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		startDate = now.AddDate(0, 0, -(weekday-1)+7)
		endDate = startDate.AddDate(0, 0, 6)

	case "2 недели вперед":
		startDate = now
		endDate = now.AddDate(0, 0, 14)

	case "Весь месяц":
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
		endDate = startDate.AddDate(0, 1, -1)

	default:
		// Парсим период вида "ДД.ММ.ГГГГ-ДД.ММ.ГГГГ"
		parts := strings.Split(messageText, "-")
		if len(parts) != 2 {
			b.sendError(chatID, "❌ Неверный формат периода. Используйте ДД.ММ.ГГГГ-ДД.ММ.ГГГГ")
			return
		}

		start, err1 := time.Parse("02.01.2006", strings.TrimSpace(parts[0]))
		end, err2 := time.Parse("02.01.2006", strings.TrimSpace(parts[1]))

		if err1 != nil || err2 != nil {
			b.sendError(chatID, "❌ Неверный формат даты. Используйте ДД.ММ.ГГГГ")
			return
		}

		if end.Before(start) {
			b.sendError(chatID, "❌ Дата окончания должна быть позже даты начала")
			return
		}

		startDate = start
		endDate = end
	}

	// Сохраняем период
	session.ScheduleStartDate = startDate
	session.ScheduleEndDate = endDate

	// Получаем тренировки на этот период
	b.showScheduleForPeriod(chatID, startDate, endDate)
}

func (b *Bot) showScheduleForDate(chatID int64, date time.Time) {
	// Получаем coachID
	user, _, _, _, err := b.UserService.GetUserProfile(chatID)
	if err != nil {
		b.sendError(chatID, "❌ Ошибка получения данных пользователя")
		return
	}

	coach, err := b.CoachService.GetCoachByUserID(user.ID)
	if err != nil {
		b.sendError(chatID, "❌ Ошибка получения данных тренера")
		return
	}

	// Получаем тренировки на эту дату
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
	end := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 0, time.Local)

	trainings, err := b.ScheduleService.GetCoachSchedule(coach.ID, start, end)
	if err != nil {
		b.sendError(chatID, "❌ Ошибка при получении расписания")
		return
	}

	// Сохраняем тренировки в сессии для отображения
	session := b.getOrCreateSession(chatID)
	session.ScheduleTrainings = trainings

	// Отображаем расписание
	b.showScheduleListView(chatID, date.Format("02.01.2006"))
}

func (b *Bot) showScheduleForPeriod(chatID int64, startDate, endDate time.Time) {
	// Получаем coachID
	user, _, _, _, err := b.UserService.GetUserProfile(chatID)
	if err != nil {
		b.sendError(chatID, "❌ Ошибка получения данных пользователя")
		return
	}

	coach, err := b.CoachService.GetCoachByUserID(user.ID)
	if err != nil {
		b.sendError(chatID, "❌ Ошибка получения данных тренера")
		return
	}

	// Получаем тренировки на период
	start := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.Local)
	end := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, time.Local)

	trainings, err := b.ScheduleService.GetCoachSchedule(coach.ID, start, end)
	if err != nil {
		b.sendError(chatID, "❌ Ошибка при получении расписания")
		return
	}

	// Сохраняем тренировки в сессии для отображения
	session := b.getOrCreateSession(chatID)
	session.ScheduleTrainings = trainings

	// Формируем описание периода
	periodDesc := fmt.Sprintf("%s - %s",
		startDate.Format("02.01.2006"),
		endDate.Format("02.01.2006"))

	// Отображаем расписание
	b.showScheduleListView(chatID, periodDesc)
}
func (b *Bot) showDateInputForSchedule(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Сегодня"),
			tgbotapi.NewKeyboardButton("Завтра"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Послезавтра"),
			tgbotapi.NewKeyboardButton("Через неделю"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("❌ Отмена"),
		),
	)

	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}

func (b *Bot) showPeriodInputForSchedule(chatID int64) {
	msg := tgbotapi.NewMessage(chatID,
		"📆 *Введите период для просмотра расписания:*\n\n"+
			"Формат: *ДД.ММ.ГГГГ-ДД.ММ.ГГГГ*\n"+
			"Пример: 15.12.2024-20.12.2024\n\n"+
			"Или выберите быстрый вариант:")
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Эта неделя"),
			tgbotapi.NewKeyboardButton("Следующая неделя"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("2 недели вперед"),
			tgbotapi.NewKeyboardButton("Весь месяц"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("❌ Отмена"),
		),
	)

	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}

// Вспомогательная функция для дня недели на русском
func getRussianDayOfWeek(day time.Weekday) string {
	days := []string{
		"Воскресенье",
		"Понедельник",
		"Вторник",
		"Среда",
		"Четверг",
		"Пятница",
		"Суббота",
	}
	if int(day) < len(days) {
		return days[day]
	}
	return ""
}

func (b *Bot) showScheduleListView(chatID int64, periodDesc string) {
	session := b.getOrCreateSession(chatID)
	trainings := session.ScheduleTrainings

	if len(trainings) == 0 {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("📭 Нет тренировок за период: %s", periodDesc))
		msg.ReplyMarkup = createScheduleManagementKeyboard()
		b.api.Send(msg)
		b.resetSession(chatID)
		return
	}

	// Группируем тренировки по дате
	trainingsByDate := make(map[string][]models.TrainingSchedule)

	for _, training := range trainings {
		dateKey := training.TrainingDate.Format("2006-01-02")
		trainingsByDate[dateKey] = append(trainingsByDate[dateKey], training)
	}

	// Сортируем даты
	var dates []string
	for date := range trainingsByDate {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	// Формируем сообщение
	var message strings.Builder
	message.WriteString(fmt.Sprintf("📅 *Расписание за период: %s*\n\n", periodDesc))

	for i, dateStr := range dates {
		date, _ := time.Parse("2006-01-02", dateStr)
		trainingsForDate := trainingsByDate[dateStr]

		// Сортируем тренировки по времени
		sort.Slice(trainingsForDate, func(i, j int) bool {
			return trainingsForDate[i].StartTime.Before(trainingsForDate[j].StartTime)
		})

		// День недели на русском
		dayOfWeek := getRussianDayOfWeek(date.Weekday())

		// Заголовок дня
		if i > 0 {
			message.WriteString("\n" + strings.Repeat("─", 30) + "\n\n")
		}

		message.WriteString(fmt.Sprintf("📅 *%s, %s*\n",
			dayOfWeek,
			date.Format("02.01.2006")))

		// Тренировки на этот день
		for _, training := range trainingsForDate {
			group, err := b.TrainingGroupService.GetGroupByID(training.GroupID)
			groupName := "Неизвестная группа"
			if err == nil && group != nil {
				groupName = group.Name
			}

			// Форматируем время
			startTime := training.StartTime.Format("15:04")
			endTime := training.EndTime.Format("15:04")

			// Описание (место и вид)
			description := "Тренировка"
			if training.Description != "" {
				description = training.Description
			}

			message.WriteString(fmt.Sprintf(
				"\n├─ 🕐 *%s-%s*\n"+
					"├─ 👥 %s\n"+
					"└─ 📍 %s\n",
				startTime,
				endTime,
				groupName,
				description,
			))
		}
	}

	// Отправляем сообщение с расписанием
	msg := tgbotapi.NewMessage(chatID, message.String())
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = createScheduleManagementKeyboard() // Возвращаемся к меню управления
	b.api.Send(msg)

	// Сбрасываем сессию
	b.resetSession(chatID)
}
