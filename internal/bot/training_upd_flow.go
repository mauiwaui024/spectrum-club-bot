package bot

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"spectrum-club-bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) handleEditTraining(chatID int64, user *models.User) {
	if user.Role != "coach" {
		b.sendError(chatID, "❌ Эта функция доступна только тренерам")
		return
	}

	session := b.getOrCreateSession(chatID)
	session.State = StateSelectingTrainingDateToEdit

	msgText := "📅 *Введите дату тренировки для редактирования:*\n\n"
	msgText += "Формат: ДД.ММ.ГГГГ\n"
	msgText += "Пример: 15.12.2024\n\n"
	msgText += "Или выберите быстрый вариант:"

	msg := tgbotapi.NewMessage(chatID, msgText)
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

func (b *Bot) handleDateSelectionForTrainingForEdit(chatID int64, messageText string) {
	session := b.getOrCreateSession(chatID)
	if session.State != StateSelectingTrainingDateToEdit {
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

	start := time.Date(
		selectedDate.Year(),
		selectedDate.Month(),
		selectedDate.Day(),
		0, 0, 0, 0, time.Local,
	)
	end := time.Date(
		selectedDate.Year(),
		selectedDate.Month(),
		selectedDate.Day(),
		23, 59, 59, 0, time.Local,
	)

	trainings, err := b.ScheduleService.GetCoachSchedule(coach.ID, start, end)
	if err != nil {
		b.sendError(chatID, "❌ Ошибка при получении расписания")
		return
	}

	if len(trainings) == 0 {
		msg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf("📭 У вас нет тренировок на %s",
				selectedDate.Format("02.01.2006")))
		msg.ReplyMarkup = createScheduleManagementKeyboard()
		b.api.Send(msg)
		b.resetSession(chatID)
		return
	}

	session.AvailableTrainingsEdit = trainings
	session.State = StateSelectingTrainingToEdit
	b.showTrainingsForEdit(chatID, trainings, selectedDate)
}

func (b *Bot) showTrainingsForEdit(chatID int64, trainings []models.TrainingSchedule, selectedDate time.Time) {
	msgText := fmt.Sprintf("📝 *Тренировки на %s:*\n\n", selectedDate.Format("02.01.2006"))

	for i, training := range trainings {
		group, _ := b.TrainingGroupService.GetGroupByID(training.GroupID)
		groupName := "Неизвестная группа"
		if group != nil {
			groupName = group.Name
		}

		dayOfWeek := getRussianDayOfWeek(training.TrainingDate.Weekday())
		msgText += fmt.Sprintf("%d. *%s*\n   🕐 %s-%s\n   👥 %s\n   📍 %s\n\n",
			i+1,
			dayOfWeek,
			training.StartTime.Format("15:04"),
			training.EndTime.Format("15:04"),
			groupName,
			training.Description,
		)
	}
	msgText += "Введите номер тренировки для редактирования или '❌ Отмена'"

	msg := tgbotapi.NewMessage(chatID, msgText)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = createCancelKeyboard()
	b.api.Send(msg)
}

func (b *Bot) handleTrainingSelectionForEdit(chatID int64, messageText string) {
	session := b.getOrCreateSession(chatID)
	if session.State != StateSelectingTrainingToEdit {
		return
	}

	if messageText == "❌ Отмена" {
		b.cancelOperation(chatID)
		return
	}

	index, err := strconv.Atoi(messageText)
	if err != nil || index < 1 || index > len(session.AvailableTrainingsEdit) {
		b.sendError(chatID, "❌ Введите корректный номер тренировки")
		return
	}

	training := session.AvailableTrainingsEdit[index-1]
	session.SelectedTrainingID = training.ID
	session.State = StateSelectingFieldToEdit

	b.showFieldSelectionMenu(chatID, &training)
}

// Упрощенное меню редактирования - только время и место
func (b *Bot) showFieldSelectionMenu(chatID int64, training *models.TrainingSchedule) {
	group, _ := b.TrainingGroupService.GetGroupByID(training.GroupID)
	groupName := "Неизвестная группа"
	if group != nil {
		groupName = group.Name
	}

	dayOfWeek := getRussianDayOfWeek(training.TrainingDate.Weekday())

	msgText := fmt.Sprintf(
		"✏️ *Редактирование тренировки*\n\n"+
			"📅 *%s, %s*\n"+
			"🕐 *Время:* %s-%s\n"+
			"👥 *Группа:* %s\n"+
			"📍 *Место:* %s\n\n"+
			"Что вы хотите сделать с тренировкой?",
		dayOfWeek,
		training.TrainingDate.Format("02.01.2006"),
		training.StartTime.Format("15:04"),
		training.EndTime.Format("15:04"),
		groupName,
		training.Description,
	)

	msg := tgbotapi.NewMessage(chatID, msgText)
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🕐 Изменить время"),
			tgbotapi.NewKeyboardButton("📍 Изменить место"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🗑️ Удалить тренировку"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("❌ Отмена"),
		),
	)

	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}

func (b *Bot) handleFieldSelectionForEdit(chatID int64, messageText string) {
	session := b.getOrCreateSession(chatID)
	if session.State != StateSelectingFieldToEdit {
		return
	}

	switch messageText {
	case "🕐 Изменить время":
		session.State = StateEditingTime
		b.showTimeEditMenu(chatID)
	case "📍 Изменить место":
		session.State = StateEditingPlace
		b.showPlaceEditMenu(chatID)
	case "🗑️ Удалить тренировку":
		session.State = StateConfirmingDeletion
		b.showDeletionTrainingConfirmation(chatID)
	case "❌ Отмена":
		b.cancelOperation(chatID)
	default:
		b.sendError(chatID, "❌ Выберите один из вариантов")
	}
}

func (b *Bot) showTimeEditMenu(chatID int64) {
	msg := tgbotapi.NewMessage(chatID,
		"🕐 *Введите новое время тренировки:*\n\n"+
			"Формат: *Начало-Конец*\n"+
			"Пример: 14:00-15:30\n\n"+
			"Или только время начала:\n"+
			"Пример: 15:00 (конец сдвинется автоматически)")

	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = createCancelKeyboard()
	b.api.Send(msg)
}

func (b *Bot) handleTimeEdit(chatID int64, messageText string) {
	session := b.getOrCreateSession(chatID)
	if session.State != StateEditingTime {
		return
	}

	if messageText == "❌ Отмена" {
		b.cancelOperation(chatID)
		return
	}

	training, err := b.ScheduleService.GetTrainingByID(session.SelectedTrainingID)
	if err != nil {
		b.sendError(chatID, "❌ Ошибка получения данных тренировки")
		return
	}

	var newStartTime, newEndTime time.Time
	duration := training.EndTime.Sub(training.StartTime)

	if strings.Contains(messageText, "-") {
		parts := strings.Split(messageText, "-")
		if len(parts) != 2 {
			b.sendError(chatID, "❌ Неверный формат. Используйте ЧЧ:ММ-ЧЧ:ММ")
			return
		}

		startTime, err := time.Parse("15:04", strings.TrimSpace(parts[0]))
		if err != nil {
			b.sendError(chatID, "❌ Неверный формат времени начала")
			return
		}

		endTime, err := time.Parse("15:04", strings.TrimSpace(parts[1]))
		if err != nil {
			b.sendError(chatID, "❌ Неверный формат времени окончания")
			return
		}

		if !endTime.After(startTime) {
			b.sendError(chatID, "❌ Время окончания должно быть позже времени начала")
			return
		}

		newStartTime = time.Date(
			training.StartTime.Year(),
			training.StartTime.Month(),
			training.StartTime.Day(),
			startTime.Hour(),
			startTime.Minute(),
			0, 0, time.Local,
		)

		newEndTime = time.Date(
			training.EndTime.Year(),
			training.EndTime.Month(),
			training.EndTime.Day(),
			endTime.Hour(),
			endTime.Minute(),
			0, 0, time.Local,
		)
	} else {
		startTime, err := time.Parse("15:04", messageText)
		if err != nil {
			b.sendError(chatID, "❌ Неверный формат времени. Используйте ЧЧ:ММ")
			return
		}

		newStartTime = time.Date(
			training.StartTime.Year(),
			training.StartTime.Month(),
			training.StartTime.Day(),
			startTime.Hour(),
			startTime.Minute(),
			0, 0, time.Local,
		)

		newEndTime = newStartTime.Add(duration)
	}

	now := time.Now()
	if isSameDay(newStartTime, now) && newStartTime.Before(now) {
		b.sendError(chatID, "❌ Нельзя перенести тренировку на прошедшее время")
		return
	}

	updates := map[string]interface{}{
		"start_time": newStartTime,
		"end_time":   newEndTime,
		// "training_date": time.Date(newStartTime.Year(), newStartTime.Month(), newStartTime.Day(), 0, 0, 0, 0, time.Local),
	}

	err = b.ScheduleService.UpdateTrainingPartial(session.SelectedTrainingID, updates)
	if err != nil {
		b.sendError(chatID, "❌ Ошибка при обновлении времени: "+err.Error())
	} else {
		msg := tgbotapi.NewMessage(chatID, "✅ Время тренировки успешно обновлено!")
		msg.ReplyMarkup = createScheduleManagementKeyboard()
		b.api.Send(msg)
	}

	b.resetSession(chatID)
}

// Редактирование места
func (b *Bot) showPlaceEditMenu(chatID int64) {
	msg := tgbotapi.NewMessage(chatID,
		"📍 *Введите новое место/описание тренировки:*\n\n"+
			"Примеры:\n"+
			"• Зал скалолазания №1\n"+
			"• Сектор боулдеринга\n"+
			"• Тренировка на выносливость")

	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = createCancelKeyboard()
	b.api.Send(msg)
}

func (b *Bot) handlePlaceEdit(chatID int64, messageText string) {
	session := b.getOrCreateSession(chatID)
	if session.State != StateEditingPlace {
		return
	}

	if messageText == "❌ Отмена" {
		b.cancelOperation(chatID)
		return
	}

	if messageText == "" {
		b.sendError(chatID, "❌ Введите новое место/описание")
		return
	}

	updates := map[string]interface{}{
		"description": messageText,
	}

	err := b.ScheduleService.UpdateTrainingPartial(session.SelectedTrainingID, updates)
	if err != nil {
		b.sendError(chatID, "❌ Ошибка при обновлении места: "+err.Error())
	} else {
		msg := tgbotapi.NewMessage(chatID, "✅ Место тренировки успешно обновлено!")
		msg.ReplyMarkup = createScheduleManagementKeyboard()
		b.api.Send(msg)
	}

	b.resetSession(chatID)
}

func isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func createCancelKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("❌ Отмена"),
		),
	)
}

func (b *Bot) showDeletionTrainingConfirmation(chatID int64) {
	session := b.getOrCreateSession(chatID)

	// Получаем тренировку для отображения информации
	training, err := b.ScheduleService.GetTrainingByID(session.SelectedTrainingID)
	if err != nil {
		b.sendError(chatID, "❌ Ошибка получения данных тренировки")
		return
	}

	group, _ := b.TrainingGroupService.GetGroupByID(training.GroupID)
	groupName := "Неизвестная группа"
	if group != nil {
		groupName = group.Name
	}

	dayOfWeek := getRussianDayOfWeek(training.TrainingDate.Weekday())

	msgText := fmt.Sprintf(
		"⚠️ *Подтвердите удаление тренировки*\n\n"+
			"Вы действительно хотите удалить эту тренировку?\n\n"+
			"📅 *%s, %s*\n"+
			"🕐 *Время:* %s-%s\n"+
			"👥 *Группа:* %s\n"+
			"📍 *Место:* %s\n\n"+
			"*Это действие невозможно отменить!*",
		dayOfWeek,
		training.TrainingDate.Format("02.01.2006"),
		training.StartTime.Format("15:04"),
		training.EndTime.Format("15:04"),
		groupName,
		training.Description,
	)

	msg := tgbotapi.NewMessage(chatID, msgText)
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("✅ Да, удалить"),
			tgbotapi.NewKeyboardButton("❌ Нет, отменить"),
		),
	)

	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}

func (b *Bot) handleDeletionConfirmation(chatID int64, messageText string) {
	session := b.getOrCreateSession(chatID)
	if session.State != StateConfirmingDeletion {
		return
	}

	switch messageText {
	case "✅ Да, удалить":
		b.deleteTraining(chatID, session)
	case "❌ Нет, отменить":
		// Возвращаемся к меню редактирования
		training, err := b.ScheduleService.GetTrainingByID(session.SelectedTrainingID)
		if err != nil {
			b.sendError(chatID, "❌ Ошибка получения данных тренировки")
			b.resetSession(chatID)
			return
		}
		session.State = StateSelectingFieldToEdit
		b.showFieldSelectionMenu(chatID, training)
	default:
		b.sendError(chatID, "❌ Пожалуйста, выберите один из вариантов")
	}
}

func (b *Bot) deleteTraining(chatID int64, session *UserSession) {
	// Получаем информацию о тренировке для сообщения
	training, err := b.ScheduleService.GetTrainingByID(session.SelectedTrainingID)
	if err != nil {
		b.sendError(chatID, "❌ Ошибка получения данных тренировки")
		return
	}

	group, _ := b.TrainingGroupService.GetGroupByID(training.GroupID)
	groupName := "Неизвестная группа"
	if group != nil {
		groupName = group.Name
	}

	// Удаляем тренировку
	err = b.ScheduleService.DeleteTraining(session.SelectedTrainingID)
	if err != nil {
		b.sendError(chatID, "❌ Ошибка при удалении тренировки: "+err.Error())
	} else {
		msg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf("✅ Тренировка успешно удалена!\n\n"+
				"📅 *%s, %s*\n"+
				"🕐 *Время:* %s-%s\n"+
				"👥 *Группа:* %s",
				getRussianDayOfWeek(training.TrainingDate.Weekday()),
				training.TrainingDate.Format("02.01.2006"),
				training.StartTime.Format("15:04"),
				training.EndTime.Format("15:04"),
				groupName))
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = createScheduleManagementKeyboard()
		b.api.Send(msg)
	}

	b.resetSession(chatID)
}
