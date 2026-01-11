package bot

import (
	"fmt"
	"spectrum-club-bot/internal/models"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// В bot/bot.go добавляем функции:

// handleCreateFromTemplates запускает процесс создания расписания из шаблонов
func (b *Bot) handleCreateFromTemplates(chatID int64, user *models.User) {
	// Проверяем права - только тренеры могут создавать расписание
	if user.Role != "coach" {
		b.sendError(chatID, "❌ Эта функция доступна только тренерам")
		return
	}

	session := b.getOrCreateSession(chatID)
	session.State = StateSelectingWeeksCount

	msg := tgbotapi.NewMessage(chatID,
		"📅 На сколько недель вперед создать расписание?\n\n"+
			"Можно ввести число от 1 до 8.\n"+
			"Рекомендуется создавать на 4 недели.")

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("1 неделя"),
			tgbotapi.NewKeyboardButton("2 недели"),
			tgbotapi.NewKeyboardButton("4 недели"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("8 недель"),
			tgbotapi.NewKeyboardButton("❌ Отмена"),
		),
	)
	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}

// handleWeeksCountSelection обрабатывает выбор количества недель
func (b *Bot) handleWeeksCountSelection(chatID int64, messageText string) {
	session := b.getOrCreateSession(chatID)
	if session.State != StateSelectingWeeksCount {
		return
	}

	if messageText == "❌ Отмена" {
		b.cancelOperation(chatID)
		return
	}

	weeksCount := 4 // по умолчанию
	switch messageText {
	case "1 неделя":
		weeksCount = 1
	case "2 недели":
		weeksCount = 2
	case "4 недели":
		weeksCount = 4
	case "8 недель":
		weeksCount = 8
	default:
		// Пробуем распарсить число
		var num int
		_, err := fmt.Sscanf(messageText, "%d", &num)
		if err != nil || num < 1 || num > 12 {
			b.sendError(chatID, "❌ Введите число от 1 до 12")
			return
		}
		weeksCount = num
	}

	session.WeeksCount = weeksCount
	session.State = StateConfirmingWeeklySchedule

	// Вычисляем даты
	weekStart := getNextMonday(time.Now())
	weekEnd := weekStart.AddDate(0, 0, weeksCount*7-1)

	msgText := fmt.Sprintf(
		"✅ Подтвердите создание расписания:\n\n"+
			"📅 Период: %s - %s\n"+
			"⏳ Недель: %d\n\n"+
			"Будут созданы тренировки для всех групп по шаблону.",
		weekStart.Format("02.01"),
		weekEnd.Format("02.01"),
		weeksCount,
	)

	msg := tgbotapi.NewMessage(chatID, msgText)
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("✅ Создать расписание"),
			tgbotapi.NewKeyboardButton("❌ Отмена"),
		),
	)
	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}

// handleWeeklyScheduleConfirmation обрабатывает подтверждение создания расписания
func (b *Bot) handleWeeklyScheduleConfirmation(chatID int64, user *models.User, messageText string) {
	session := b.getOrCreateSession(chatID)
	if session.State != StateConfirmingWeeklySchedule {
		return
	}

	switch messageText {
	case "✅ Создать расписание":
		b.createWeeklySchedule(chatID, user)
	case "❌ Отмена":
		b.cancelOperation(chatID)
	default:
		b.sendError(chatID, "❌ Неизвестная команда")
	}
}

// createWeeklySchedule создает расписание на недели вперед
func (b *Bot) createWeeklySchedule(chatID int64, user *models.User) {
	session := b.getOrCreateSession(chatID)

	// Получаем данные тренера
	coach, err := b.CoachService.GetCoachByUserID(user.ID)
	if err != nil {
		b.sendError(chatID, "❌ Ошибка получения данных тренера")
		b.resetSession(chatID)
		return
	}

	// Дата начала (ближайший понедельник)
	weekStart := getNextMonday(time.Now())

	// Показываем сообщение о начале процесса
	msg := tgbotapi.NewMessage(chatID, "⏳ Создаю расписание... Это может занять несколько секунд.")
	b.api.Send(msg)

	// Используем метод сервиса (нужно будет добавить его в интерфейс)
	createdCount, err := b.ScheduleService.CreateTrainingsFromTemplates(
		weekStart,
		coach.ID,
		user.ID,
		session.WeeksCount,
	)

	if err != nil {
		b.sendError(chatID, "❌ Ошибка создания расписания: "+err.Error())
		b.resetSession(chatID)
		return
	}

	// Формируем результат
	weekEnd := weekStart.AddDate(0, 0, session.WeeksCount*7-1)
	msgText := fmt.Sprintf(
		"✅ Расписание успешно создано!\n\n"+
			"📊 Создано тренировок: %d\n"+
			"📅 Период: %s - %s\n\n"+
			"Расписание создано на %d недель вперед.",
		createdCount,
		weekStart.Format("02.01"),
		weekEnd.Format("02.01"),
		session.WeeksCount,
	)

	msg = tgbotapi.NewMessage(chatID, msgText)
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	b.api.Send(msg)

	b.resetSession(chatID)
}

// Вспомогательная функция для получения ближайшего понедельника
func getNextMonday(from time.Time) time.Time {
	daysUntilMonday := (8 - int(from.Weekday())) % 7
	if daysUntilMonday == 0 {
		daysUntilMonday = 7
	}
	return from.AddDate(0, 0, daysUntilMonday)
}
