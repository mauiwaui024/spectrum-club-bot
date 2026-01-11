package bot

import (
	"fmt"
	"spectrum-club-bot/internal/models"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) handleAddTraining(chatID int64, user *models.User) {
	// Проверяем права - только тренеры могут добавлять тренировки
	if user.Role != "coach" {
		b.sendError(chatID, "❌ Эта функция доступна только тренерам")
		return
	}

	session := b.getOrCreateSession(chatID)
	session.State = StateSelectingGroup

	// Получаем список групп
	groups, err := b.TrainingGroupService.GetAllGroups()
	if err != nil {
		b.sendError(chatID, "❌ Ошибка при получении списка групп")
		b.resetSession(chatID)
		return
	}

	msg := tgbotapi.NewMessage(chatID, "👥 Выберите группу для тренировки:")
	keyboard := b.createGroupsKeyboard(groups)
	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}

func (b *Bot) handleGroupSelection(chatID int64, messageText string) {
	session := b.getOrCreateSession(chatID)
	if session.State != StateSelectingGroup {
		return
	}

	if messageText == "❌ Отмена" {
		b.cancelOperation(chatID)
		return
	}

	// Ищем группу по имени
	groups, err := b.TrainingGroupService.GetAllGroups()
	if err != nil {
		b.sendError(chatID, "❌ Ошибка при поиске группы")
		return
	}

	var selectedGroup *models.TrainingGroup
	for i := range groups {
		if messageText == groups[i].Name {
			selectedGroup = &groups[i]
			break
		}
	}

	if selectedGroup == nil {
		b.sendError(chatID, "❌ Группа не найдена")
		return
	}

	session.SelectedGroupID = selectedGroup.ID
	session.State = StateSelectingDate

	// Показываем выбор даты
	b.showDateSelection(chatID)
}
func (b *Bot) showDateSelection(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "📅 Выберите дату тренировки:\n\nМожно ввести дату в формате ДД.ММ.ГГГГ (например: 15.12.2024)")

	// Создаем клавиатуру с быстрыми вариантами
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Сегодня"),
			tgbotapi.NewKeyboardButton("Завтра"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("❌ Отмена"),
		),
	)

	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}

func (b *Bot) handleDateSelection(chatID int64, messageText string) {
	session := b.getOrCreateSession(chatID)
	if session.State != StateSelectingDate {
		return
	}

	if messageText == "❌ Отмена" {
		b.cancelOperation(chatID)
		return
	}

	var selectedDate time.Time
	now := time.Now()

	switch messageText {
	case "Сегодня":
		selectedDate = now
	case "Завтра":
		selectedDate = now.AddDate(0, 0, 1)
	default:
		// Парсим дату из текста
		parsedDate, err := time.Parse("02.01.2006", messageText)
		if err != nil {
			b.sendError(chatID, "❌ Неверный формат даты. Используйте ДД.ММ.ГГГГ")
			return
		}
		selectedDate = parsedDate
	}

	// Проверяем что дата не в прошлом
	if selectedDate.YearDay() < now.YearDay() && selectedDate.Year() <= now.Year() {
		b.sendError(chatID, "❌ Нельзя создать тренировку в прошлом")
		return
	}

	session.SelectedDate = selectedDate
	session.State = StateSelectingTime

	b.showTimeSelection(chatID)
}

func (b *Bot) showTimeSelection(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "⏰ Выберите время начала тренировки:\n\nМожно ввести время в формате ЧЧ:ММ (например: 14:30)")

	// Создаем клавиатуру с быстрыми вариантами времени
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("09:00"),
			tgbotapi.NewKeyboardButton("10:00"),
			tgbotapi.NewKeyboardButton("11:00"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("14:00"),
			tgbotapi.NewKeyboardButton("15:00"),
			tgbotapi.NewKeyboardButton("16:00"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("17:00"),
			tgbotapi.NewKeyboardButton("18:00"),
			tgbotapi.NewKeyboardButton("19:00"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("❌ Отмена"),
		),
	)

	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}

func (b *Bot) handleTimeSelection(chatID int64, messageText string) {
	session := b.getOrCreateSession(chatID)
	if session.State != StateSelectingTime {
		return
	}

	if messageText == "❌ Отмена" {
		b.cancelOperation(chatID)
		return
	}

	// Парсим время
	startTime, err := time.Parse("15:04", messageText)
	if err != nil {
		b.sendError(chatID, "❌ Неверный формат времени. Используйте ЧЧ:ММ")
		return
	}

	// Комбинируем дату и время
	combinedDateTime := time.Date(
		session.SelectedDate.Year(),
		session.SelectedDate.Month(),
		session.SelectedDate.Day(),
		startTime.Hour(),
		startTime.Minute(),
		0, 0, time.Local,
	)

	// Проверяем что время не в прошлом
	if combinedDateTime.Before(time.Now()) {
		b.sendError(chatID, "❌ Нельзя создать тренировку в прошлом")
		return
	}

	session.SelectedStartTime = combinedDateTime
	session.State = StateSelectingDuration

	b.showDurationSelection(chatID)
}

func (b *Bot) showDurationSelection(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "⏱️ Выберите продолжительность тренировки:")

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("45 минут"),
			tgbotapi.NewKeyboardButton("1 час"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("1.5 часа"),
			tgbotapi.NewKeyboardButton("2 часа"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("❌ Отмена"),
		),
	)

	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}

func (b *Bot) handleDurationSelection(chatID int64, messageText string) {
	session := b.getOrCreateSession(chatID)
	if session.State != StateSelectingDuration {
		return
	}

	if messageText == "❌ Отмена" {
		b.cancelOperation(chatID)
		return
	}

	var duration time.Duration
	switch messageText {
	case "45 минут":
		duration = 45 * time.Minute
	case "1 час":
		duration = time.Hour
	case "1.5 часа":
		duration = 90 * time.Minute
	case "2 часа":
		duration = 2 * time.Hour
	default:
		b.sendError(chatID, "❌ Неизвестная продолжительность")
		return
	}

	session.SelectedDuration = duration
	session.State = StateConfirmingTraining

	b.showTrainingConfirmation(chatID)
}

func (b *Bot) showTrainingConfirmation(chatID int64) {
	session := b.getOrCreateSession(chatID)

	group, err := b.TrainingGroupService.GetGroupByID(session.SelectedGroupID)
	if err != nil {
		b.sendError(chatID, "❌ Ошибка получения данных группы")
		return
	}

	endTime := session.SelectedStartTime.Add(session.SelectedDuration)

	msgText := fmt.Sprintf(
		"✅ Подтвердите создание тренировки:\n\n"+
			"👥 Группа: %s\n"+
			"📅 Дата: %s\n"+
			"⏰ Время: %s - %s\n"+
			"⏱️ Продолжительность: %s",
		group.Name,
		session.SelectedStartTime.Format("02.01.2006"),
		session.SelectedStartTime.Format("15:04"),
		endTime.Format("15:04"),
		formatDuration(session.SelectedDuration),
	)

	msg := tgbotapi.NewMessage(chatID, msgText)

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("✅ Создать тренировку"),
			tgbotapi.NewKeyboardButton("❌ Отмена"),
		),
	)

	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	if hours == 0 {
		return fmt.Sprintf("%d минут", minutes)
	}
	if minutes == 0 {
		return fmt.Sprintf("%d часов", hours)
	}
	return fmt.Sprintf("%d ч %d мин", hours, minutes)
}

func (b *Bot) handleTrainingConfirmation(chatID int64, messageText string) {
	session := b.getOrCreateSession(chatID)
	if session.State != StateConfirmingTraining {
		return
	}

	switch messageText {
	case "✅ Создать тренировку":
		b.createTraining(chatID, session)
	case "❌ Отмена":
		b.cancelOperation(chatID)
	default:
		b.sendError(chatID, "❌ Неизвестная команда")
	}
}

func (b *Bot) createTraining(chatID int64, session *UserSession) {
	// Получаем coachID из пользователя
	user, _, _, _, err := b.UserService.GetUserProfile(chatID)
	if err != nil {
		b.sendError(chatID, "❌ Ошибка получения данных тренера")
		return
	}

	coach, err := b.CoachService.GetCoachByUserID(user.ID)
	if err != nil {
		b.sendError(chatID, "❌ Ошибка получения данных тренера")
		return
	}

	endTime := session.SelectedStartTime.Add(session.SelectedDuration)

	training := &models.TrainingSchedule{
		GroupID:      session.SelectedGroupID,
		CoachID:      &coach.ID,
		TrainingDate: session.SelectedStartTime,
		StartTime:    session.SelectedStartTime,
		EndTime:      endTime,
		Description:  session.TrainingDescription,
		CreatedBy:    &user.ID,
	}

	err = b.ScheduleService.CreateTraining(training)
	if err != nil {
		b.sendError(chatID, "❌ Ошибка при создании тренировки: "+err.Error())
	} else {
		msg := tgbotapi.NewMessage(chatID, "✅ Тренировка успешно создана!")
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		b.api.Send(msg)
	}

	b.resetSession(chatID)
}
