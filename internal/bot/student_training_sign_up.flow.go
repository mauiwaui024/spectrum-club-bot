package bot

import (
	"fmt"
	"strconv"
	"time"

	"spectrum-club-bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Флоу записи на тренировку
func (b *Bot) handleSignUpForTraining(chatID int64, user *models.User) {
	if user.Role != "student" {
		b.sendError(chatID, "❌ Эта функция доступна только ученикам")
		return
	}

	// Получаем студента по user_id
	student, err := b.StudentService.GetStudentByUserID(user.ID)
	if err != nil {
		b.sendError(chatID, "❌ Ошибка получения данных студента")
		return
	}

	session := b.getOrCreateSession(chatID)
	session.SelectedStudentForSignUpID = int(student.ID)
	session.State = StateSelectingTrainingDateToSignUp

	msgText := "📅 *Выберите дату для записи на тренировку:*\n\n"
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

func (b *Bot) handleDateSelectionForTrainingSignUp(chatID int64, messageText string) {
	session := b.getOrCreateSession(chatID)
	if session.State != StateSelectingTrainingDateToSignUp {
		return
	}

	if messageText == "❌ Отмена" {
		b.sendWelcomeMessage(chatID, nil)
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

	// Получаем тренировки на выбранную дату
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

	// Получаем доступные тренировки
	trainings, err := b.ScheduleService.GetTrainingsByDateRange(start, end)
	if err != nil {
		b.sendError(chatID, "❌ Ошибка при получении расписания")
		return
	}

	// Фильтруем тренировки: только те, которые еще не начались
	var availableTrainings []models.TrainingSchedule
	nowTime := time.Now()

	for _, training := range trainings {
		// Тренировка должна быть в будущем
		if training.TrainingDate.After(nowTime) {
			// Проверяем, есть ли у студента активный абонемент
			activeSub, err := b.SubscriptionService.GetActiveSubscription(int64(session.SelectedStudentForSignUpID))
			if err == nil && activeSub != nil && activeSub.RemainingLessons > 0 {
				// Проверяем, не записан ли уже студент на эту тренировку
				existing, err := b.AttendanceService.GetStudentAttendanceForTraining(session.SelectedStudentForSignUpID, training.ID)
				if err == nil && existing == nil {
					availableTrainings = append(availableTrainings, training)
				}
			}
		}
	}

	if len(availableTrainings) == 0 {
		msg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf("📭 Нет доступных тренировок для записи на %s\n\nПроверьте:\n• Есть ли у вас активный абонемент\n• Не закончились ли занятия\n• Возможно, вы уже записаны на все тренировки в этот день",
				selectedDate.Format("02.01.2006")))
		msg.ReplyMarkup = createStudentMainKeyboard()
		b.api.Send(msg)
		b.resetSession(chatID)
		return
	}

	session.AvailableTrainingsForSignUp = availableTrainings
	session.State = StateSelectingTrainingToSignUp

	b.showAvailableTrainingsForSignUp(chatID, availableTrainings, selectedDate)
}

func (b *Bot) showAvailableTrainingsForSignUp(chatID int64, trainings []models.TrainingSchedule, selectedDate time.Time) {
	msgText := fmt.Sprintf("📝 *Доступные тренировки на %s:*\n\n", selectedDate.Format("02.01.2006"))

	for i, training := range trainings {
		group, _ := b.TrainingGroupService.GetGroupByID(training.GroupID)
		groupName := "Неизвестная группа"
		if group != nil {
			groupName = group.Name
		}

		coachName := "Неизвестный тренер"
		if training.CoachID != nil {
			coach, _ := b.CoachService.GetByCoachID(*training.CoachID)
			if coach != nil {
				user, _ := b.UserService.GetByID(coach.UserID)
				if user != nil {
					coachName = user.FirstName + " " + user.LastName
				}
			}
		}

		// Текущее количество записанных
		currentCount, _, _, _ := b.AttendanceService.GetTrainingStats(training.ID)
		maxCount := "без ограничений"
		if training.MaxParticipants != nil {
			maxCount = fmt.Sprintf("%d/%d", currentCount, *training.MaxParticipants)
		}

		dayOfWeek := getRussianDayOfWeek(training.TrainingDate.Weekday())
		msgText += fmt.Sprintf("%d. *%s*\n   🕐 %s-%s\n   👥 %s\n   🏋️ %s\n   📍 %s\n   👥 Места: %s\n\n",
			i+1,
			dayOfWeek,
			training.StartTime.Format("15:04"),
			training.EndTime.Format("15:04"),
			groupName,
			coachName,
			training.Description,
			maxCount,
		)
	}
	msgText += "Введите номер тренировки для записи или отправьте '❌ Отмена'"

	msg := tgbotapi.NewMessage(chatID, msgText)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = createCancelKeyboard()
	b.api.Send(msg)
}

func (b *Bot) handleTrainingSelectionForSignUp(chatID int64, messageText string) {
	session := b.getOrCreateSession(chatID)
	if session.State != StateSelectingTrainingToSignUp {
		return
	}

	if messageText == "❌ Отмена" {
		b.sendWelcomeMessage(chatID, nil)
		b.resetSession(chatID)
		return
	}

	index, err := strconv.Atoi(messageText)
	if err != nil || index < 1 || index > len(session.AvailableTrainingsForSignUp) {
		b.sendError(chatID, "❌ Введите корректный номер тренировки")
		return
	}

	training := session.AvailableTrainingsForSignUp[index-1]
	session.SelectedTrainingForSignUpID = training.ID
	session.State = StateConfirmingTrainingSignUp

	b.showTrainingSignUpConfirmation(chatID, training)
}

func (b *Bot) showTrainingSignUpConfirmation(chatID int64, training models.TrainingSchedule) {
	// Получаем информацию о тренировке
	group, _ := b.TrainingGroupService.GetGroupByID(training.GroupID)
	groupName := "Неизвестная группа"
	if group != nil {
		groupName = group.Name
	}
	coachName := "Неизвестный тренер"
	if training.CoachID != nil {
		coach, _ := b.CoachService.GetByCoachID(*training.CoachID)
		if coach != nil {
			user, _ := b.UserService.GetByID(coach.UserID)
			if user != nil {
				coachName = user.FirstName + " " + user.LastName
			}
		}
	}

	dayOfWeek := getRussianDayOfWeek(training.TrainingDate.Weekday())

	// Проверяем абонемент
	session := b.getOrCreateSession(chatID)
	activeSub, _ := b.SubscriptionService.GetActiveSubscription(int64(session.SelectedStudentForSignUpID))

	msgText := fmt.Sprintf(
		"✅ *Подтвердите запись на тренировку:*\n\n"+
			"📅 *%s, %s*\n"+
			"🕐 *Время:* %s-%s\n"+
			"👥 *Группа:* %s\n"+
			"🏋️ *Тренер:* %s\n"+
			"📍 *Место:* %s\n\n",
		dayOfWeek,
		training.TrainingDate.Format("02.01.2006"),
		training.StartTime.Format("15:04"),
		training.EndTime.Format("15:04"),
		groupName,
		coachName,
		training.Description,
	)

	if activeSub != nil {
		msgText += fmt.Sprintf("Ваш абонемент: %d/%d занятий осталось\n\n",
			activeSub.RemainingLessons, activeSub.TotalLessons)
	}

	msgText += "Вы уверены, что хотите записаться?"

	msg := tgbotapi.NewMessage(chatID, msgText)
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("✅ Да, записаться"),
			tgbotapi.NewKeyboardButton("❌ Нет, отменить"),
		),
	)

	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}

func (b *Bot) handleTrainingSignUpConfirmation(chatID int64, messageText string) {
	session := b.getOrCreateSession(chatID)
	if session.State != StateConfirmingTrainingSignUp {
		return
	}

	switch messageText {
	case "✅ Да, записаться":
		b.processTrainingSignUp(chatID, session)
	case "❌ Нет, отменить":
		b.sendWelcomeMessage(chatID, nil)
		b.resetSession(chatID)
	default:
		b.sendError(chatID, "❌ Пожалуйста, выберите один из вариантов")
	}
}

func (b *Bot) processTrainingSignUp(chatID int64, session *UserSession) {
	// Записываем студента на тренировку
	err := b.AttendanceService.SignUpForTraining(session.SelectedStudentForSignUpID, session.SelectedTrainingForSignUpID)
	if err != nil {
		b.sendError(chatID, "❌ Ошибка при записи: "+err.Error())
	} else {
		// Получаем информацию о тренировке для сообщения
		training, _ := b.ScheduleService.GetTrainingByID(session.SelectedTrainingForSignUpID)
		if training != nil {
			group, _ := b.TrainingGroupService.GetGroupByID(training.GroupID)
			groupName := "Неизвестная группа"
			if group != nil {
				groupName = group.Name
			}

			msgText := fmt.Sprintf(
				"✅ *Вы успешно записаны на тренировку!*\n\n"+
					"📅 *Дата:* %s\n"+
					"🕐 *Время:* %s-%s\n"+
					"👥 *Группа:* %s\n"+
					"📍 *Место:* %s\n\n"+
					"Не забудьте прийти за 10 минут до начала!",
				training.TrainingDate.Format("02.01.2006"),
				training.StartTime.Format("15:04"),
				training.EndTime.Format("15:04"),
				groupName,
				training.Description,
			)

			msg := tgbotapi.NewMessage(chatID, msgText)
			msg.ParseMode = "Markdown"
			msg.ReplyMarkup = createStudentMainKeyboard()
			b.api.Send(msg)
		} else {
			msg := tgbotapi.NewMessage(chatID, "✅ Вы успешно записаны на тренировку!")
			msg.ReplyMarkup = createStudentMainKeyboard()
			b.api.Send(msg)
		}
	}

	b.resetSession(chatID)
}

func (b *Bot) handleMyRegistrations(chatID int64, user *models.User) {
	if user.Role != "student" {
		b.sendError(chatID, "❌ Эта функция доступна только ученикам")
		return
	}

	// Получаем студента
	student, err := b.StudentService.GetStudentByUserID(user.ID)
	if err != nil {
		b.sendError(chatID, "❌ Ошибка получения данных студента")
		return
	}

	// Получаем записи на тренировки за последние 30 дней и будущие
	start := time.Now().AddDate(0, 0, -30)
	end := time.Now().AddDate(0, 0, 30)

	attendances, err := b.AttendanceService.GetAttendanceByStudent(int(student.ID), start, end)
	if err != nil {
		b.sendError(chatID, "❌ Ошибка при получении записей")
		return
	}

	if len(attendances) == 0 {
		msg := tgbotapi.NewMessage(chatID, "📭 У вас нет записей на тренировки.")
		msg.ReplyMarkup = createStudentMainKeyboard()
		b.api.Send(msg)
		return
	}

	// Группируем записи по дате и статусу
	var upcomingTrainings []models.Attendance
	var pastTrainings []models.Attendance

	now := time.Now()
	for _, attendance := range attendances {
		training, err := b.ScheduleService.GetTrainingByID(attendance.TrainingID)
		if err != nil {
			continue
		}

		if training.StartTime.After(now) {
			upcomingTrainings = append(upcomingTrainings, attendance)
		} else {
			pastTrainings = append(pastTrainings, attendance)
		}
	}

	// Формируем сообщение
	var message string
	if len(upcomingTrainings) > 0 {
		message += "📅 *Предстоящие тренировки:*\n\n"
		for i, attendance := range upcomingTrainings {
			training, _ := b.ScheduleService.GetTrainingByID(attendance.TrainingID)
			group, _ := b.TrainingGroupService.GetGroupByID(training.GroupID)
			groupName := "Неизвестная группа"
			if group != nil {
				groupName = group.Name
			}

			dayOfWeek := getRussianDayOfWeek(training.TrainingDate.Weekday())
			status := "✅ Записан"
			if attendance.Attended {
				status = "✅ Посещена"
			}

			message += fmt.Sprintf("%d. *%s, %s*\n   🕐 %s-%s\n   👥 %s\n   📍 %s\n   %s\n\n",
				i+1,
				dayOfWeek,
				training.TrainingDate.Format("02.01.2006"),
				training.StartTime.Format("15:04"),
				training.EndTime.Format("15:04"),
				groupName,
				training.Description,
				status,
			)
		}
	}

	if len(pastTrainings) > 0 {
		message += "\n📊 *Прошедшие тренировки:*\n\n"
		for i, attendance := range pastTrainings {
			training, _ := b.ScheduleService.GetTrainingByID(attendance.TrainingID)
			group, _ := b.TrainingGroupService.GetGroupByID(training.GroupID)
			groupName := "Неизвестная группа"
			if group != nil {
				groupName = group.Name
			}

			dayOfWeek := getRussianDayOfWeek(training.TrainingDate.Weekday())
			status := "❌ Пропущена"
			if attendance.Attended {
				status = "✅ Посещена"
			}

			message += fmt.Sprintf("%d. *%s, %s*\n   🕐 %s-%s\n   👥 %s\n   %s\n",
				i+1,
				dayOfWeek,
				training.TrainingDate.Format("02.01.2006"),
				training.StartTime.Format("15:04"),
				training.EndTime.Format("15:04"),
				groupName,
				status,
			)

			if attendance.Notes != "" {
				message += fmt.Sprintf("   📝 Заметки: %s\n", attendance.Notes)
			}
			message += "\n"
		}
	}

	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = createStudentMainKeyboard()
	b.api.Send(msg)
}

func (b *Bot) handleMySubscription(chatID int64, user *models.User) {
	if user.Role != "student" {
		b.sendError(chatID, "❌ Эта функция доступна только ученикам")
		return
	}

	// Получаем студента
	student, err := b.StudentService.GetStudentByUserID(user.ID)
	if err != nil {
		b.sendError(chatID, "❌ Ошибка получения данных студента")
		return
	}

	// Получаем активные абонементы
	subscriptions, err := b.SubscriptionService.GetSubscriptionsByStudentID(student.ID)
	if err != nil {
		b.sendError(chatID, "❌ Ошибка при получении данных абонемента")
		return
	}

	var activeSubscriptions []*models.Subscription
	var expiredSubscriptions []*models.Subscription
	now := time.Now()

	for _, subscription := range subscriptions {
		if subscription.RemainingLessons > 0 && subscription.EndDate.After(now) {
			activeSubscriptions = append(activeSubscriptions, subscription)
		} else {
			expiredSubscriptions = append(expiredSubscriptions, subscription)
		}
	}

	// Формируем сообщение
	msgText := "🎫 *Мой абонемент*\n\n"

	if len(activeSubscriptions) == 0 && len(expiredSubscriptions) == 0 {
		msgText += "У вас нет активных абонементов.\n\nОбратитесь к тренеру для приобретения абонемента."
	} else {
		if len(activeSubscriptions) > 0 {
			msgText += "✅ *Активные абонементы:*\n\n"
			for i, sub := range activeSubscriptions {
				msgText += fmt.Sprintf("%d. *%d/%d занятий*\n", i+1, sub.RemainingLessons, sub.TotalLessons)
				msgText += fmt.Sprintf("   📅 Действует до: %s\n", sub.EndDate.Format("02.01.2006"))
				// msgText += fmt.Sprintf("   🏷️ Тип: %s\n", sub.SubscriptionType)
				msgText += "\n"
			}
		}

		if len(expiredSubscriptions) > 0 {
			msgText += "⏰ *История абонементов:*\n\n"
			for i, sub := range expiredSubscriptions {
				status := "🔄 Использован"
				if sub.RemainingLessons > 0 && sub.EndDate.Before(now) {
					status = "⏰ Истек"
				}

				msgText += fmt.Sprintf("%d. %s\n", i+1, status)
				msgText += fmt.Sprintf("   📊 %d/%d занятий\n", sub.TotalLessons-sub.RemainingLessons, sub.TotalLessons)
				// msgText += fmt.Sprintf("   🏷️ Тип: %s\n", sub.SubscriptionType)
				msgText += "\n"
			}
		}
	}

	msg := tgbotapi.NewMessage(chatID, msgText)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = createStudentMainKeyboard()
	b.api.Send(msg)
}
