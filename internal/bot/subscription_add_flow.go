package bot

import (
	"fmt"
	"spectrum-club-bot/internal/models"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) getOrCreateSession(chatID int64) *UserSession {
	b.mu.Lock()
	defer b.mu.Unlock()

	if session, exists := b.userSessions[chatID]; exists {
		return session
	}

	session := &UserSession{State: StateDefault}
	b.userSessions[chatID] = session
	return session
}

func (b *Bot) handleAddSubscription(chatID int64) {
	user, _, _, _, err := b.UserService.GetUserProfile(chatID)
	if err != nil || user == nil || user.Role != "coach" {
		b.sendError(chatID, "❌ Эта функция доступна только тренерам")
		return
	}

	students, err := b.UserService.GetAllStudents()
	if err != nil {
		b.sendError(chatID, "❌ Ошибка при получении списка учеников")
		return
	}

	if len(students) == 0 {
		msg := tgbotapi.NewMessage(chatID, "📭 Список учеников пуст")
		b.sendTelegram(msg)
		return
	}

	if err := b.showStudentsForSelection(chatID, students); err != nil {
		b.sendError(chatID, "❌ Не удалось показать список учеников")
		return
	}

	session := b.getOrCreateSession(chatID)
	session.State = StateSelectingStudent
	session.StudentsForSelection = students
}

func (b *Bot) showStudentsForSelection(chatID int64, students []*models.User) error {
	msgText := "Выберите ученика:\n\n"
	for i, student := range students {
		displayName := getStudentDisplayName(student)
		msgText += fmt.Sprintf("%d. %s\n", i+1, displayName)
	}
	msgText += "\nВведите номер ученика или нажмите «❌ Отмена»"

	return b.sendLongPlain(chatID, msgText, createCancelKeyboard())
}

// Вспомогательная функция для получения отображаемого имени ученика
func getStudentDisplayName(student *models.User) string {
	if student.FirstName != "" && student.LastName != "" {
		return fmt.Sprintf("%s %s", student.FirstName, student.LastName)
	} else if student.FirstName != "" {
		return student.FirstName
	} else if student.Username != "" {
		return "@" + student.Username
	}
	return "Неизвестный ученик"
}

func (b *Bot) handleStudentSelection(chatID int64, messageText string) {
	session := b.getOrCreateSession(chatID)
	if session.State != StateSelectingStudent {
		return
	}

	if isCancelText(messageText) {
		b.cancelOperation(chatID, nil)
		return
	}

	// Парсим номер ученика
	index, err := strconv.Atoi(messageText)
	if err != nil || index < 1 || index > len(session.StudentsForSelection) {
		b.sendError(chatID, "❌ Пожалуйста, введите корректный номер ученика")
		return
	}

	// Получаем выбранного ученика по индексу
	selectedStudent := session.StudentsForSelection[index-1]
	session.SelectedStudentID = selectedStudent.ID
	session.State = StateSelectingSubscriptionType

	// Показываем выбор типа абонемента
	b.showSubscriptionTypes(chatID, selectedStudent)
}

func (b *Bot) showSubscriptionTypes(chatID int64, student *models.User) {
	msg := tgbotapi.NewMessage(chatID,
		fmt.Sprintf("🎫 Выберите тип абонемента для %s %s:",
			student.FirstName, student.LastName))

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("⛰️Пробное занятие"),
			tgbotapi.NewKeyboardButton("💪 Абонеменет на 12 занятий\n(Несгораемый)"),
			tgbotapi.NewKeyboardButton("⛏️ Абонемент на 16 занятий\n(30дней)"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("❌ Отмена"),
		),
	)

	msg.ReplyMarkup = keyboard
	b.sendTelegram(msg)
}

func (b *Bot) handleSubscriptionTypeSelection(chatID int64, messageText string) {
	session := b.getOrCreateSession(chatID)
	if session.State != StateSelectingSubscriptionType {
		return
	}

	if isCancelText(messageText) {
		b.cancelOperation(chatID, nil)
		return
	}

	// Маппинг текста к типам абонементов
	subscriptionMap := map[string]string{
		"⛰️Пробное занятие":                         "1",
		"💪 Абонеменет на 12 занятий\n(Несгораемый)": "12",
		"⛏️ Абонемент на 16 занятий\n(30дней)":      "16",
	}

	subscriptionType, exists := subscriptionMap[messageText]
	if !exists {
		b.sendError(chatID, "❌ Неизвестный тип абонемента")
		return
	}

	session.SelectedSubscriptionType = subscriptionType
	session.State = StateConfirming

	b.showConfirmation(chatID)
}

func (b *Bot) showConfirmation(chatID int64) {
	session := b.getOrCreateSession(chatID)

	students, _ := b.UserService.GetAllStudents()
	var studentName string
	for _, s := range students {
		if s.ID == session.SelectedStudentID {
			studentName = fmt.Sprintf("%s %s", s.FirstName, s.LastName)
			break
		}
	}

	subscriptionNames := map[string]string{
		"1":  "⛰️Пробное занятие",
		"12": "💪 Абонеменет на 12 занятий\n(Несгораемый)",
		"16": "⛏️ Абонемент на 16 занятий\n(30дней)",
	}

	msg := tgbotapi.NewMessage(chatID,
		fmt.Sprintf("✅ Подтвердите добавление:\n\n👤 Ученик: %s\n🎫 Абонемент: %s",
			studentName, subscriptionNames[session.SelectedSubscriptionType]))

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("✅ Подтвердить"),
			tgbotapi.NewKeyboardButton("❌ Отмена"),
		),
	)

	msg.ReplyMarkup = keyboard
	b.sendTelegram(msg)
}

func (b *Bot) handleConfirmation(chatID int64, messageText string) {
	session := b.getOrCreateSession(chatID)
	if session.State != StateConfirming {
		return
	}

	if isCancelText(messageText) {
		b.cancelOperation(chatID, nil)
		return
	}

	switch messageText {
	case "✅ Подтвердить":
		b.addSubscription(chatID, session)
	default:
		b.sendError(chatID, "❌ Неизвестная команда")
	}
}

func (b *Bot) addSubscription(chatID int64, session *UserSession) {
	// var err error
	studentFromStudents, err := b.StudentService.GetStudentByUserID(session.SelectedStudentID)
	if err != nil {
		b.sendError(chatID, "❌ Ошибка при получении данных студента: "+err.Error())
		b.resetSession(chatID)
		return
	}

	// Получаем информацию о типе абонемента для уведомления
	subscriptionNames := map[string]string{
		"1":  "⛰️ Пробное занятие (1 занятие на 30 дней)",
		"12": "💪 Абонемент на 12 занятий (Несгораемый)",
		"16": "⛏️ Абонемент на 16 занятий (30 дней)",
	}
	subscriptionName := subscriptionNames[session.SelectedSubscriptionType]
	if subscriptionName == "" {
		subscriptionName = "Абонемент"
	}

	var totalLessons int
	switch session.SelectedSubscriptionType {
	case "1":
		err = b.SubscriptionService.Create1For30Days(studentFromStudents.ID)
		totalLessons = 1
	case "12":
		err = b.SubscriptionService.Create12Unlimited(studentFromStudents.ID)
		totalLessons = 12
	case "16":
		err = b.SubscriptionService.Create16For30Days(studentFromStudents.ID)
		totalLessons = 16
	}

	if err != nil {
		b.sendError(chatID, "❌ Ошибка при добавлении абонемента: "+err.Error())
		b.resetSession(chatID)
	} else {
		b.showMainKeyboardAfterOperation(chatID, "✅ Абонемент успешно добавлен!")
		b.resetSession(chatID)

		// Отправляем уведомление студенту
		studentUser, err := b.UserService.GetByID(session.SelectedStudentID)
		if err == nil && studentUser != nil {
			msgText := fmt.Sprintf(
				"🎫 Вам зачислен абонемент!\n\n"+
					"📋 Тип: %s\n"+
					"📊 Количество занятий: %d\n\n"+
					"Теперь вы можете записываться на тренировки!",
				subscriptionName,
				totalLessons,
			)
			msg := tgbotapi.NewMessage(studentUser.TelegramID, msgText)
			b.sendTelegram(msg)
		}
	}
}

/////////////////////

func (b *Bot) resetSession(chatID int64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.userSessions, chatID)
}

func (b *Bot) cancelOperation(chatID int64, user *models.User) {
	if user != nil {
		b.sendMainMenuWithRefresh(chatID, user, "❌ Операция отменена")
		b.resetSession(chatID)
		return
	}

	// Если пользователь не передан, получаем его
	userProfile, _, _, _, err := b.UserService.GetUserProfile(chatID)
	if err == nil && userProfile != nil {
		b.sendMainMenuWithRefresh(chatID, userProfile, "❌ Операция отменена")
	} else {
		msg := tgbotapi.NewMessage(chatID, "❌ Операция отменена")
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		b.sendTelegram(msg)
	}
	b.resetSession(chatID)
}

func isCancelText(text string) bool {
	normalized := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(text), " ", ""))
	return strings.Contains(normalized, "отмена")
}

// showMainKeyboardAfterOperation показывает главную клавиатуру после завершения операции
func (b *Bot) showMainKeyboardAfterOperation(chatID int64, message string) {
	user, _, _, _, err := b.UserService.GetUserProfile(chatID)
	if err == nil && user != nil {
		b.sendMainMenuWithRefresh(chatID, user, message)
	} else {
		msg := tgbotapi.NewMessage(chatID, message)
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		b.sendTelegram(msg)
	}
}

func (b *Bot) sendError(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	b.sendTelegram(msg)
}
