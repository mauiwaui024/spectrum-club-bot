package bot

import (
	"fmt"
	"spectrum-club-bot/internal/models"
	"strconv"

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
	// ... существующий код до показа учеников ...

	students, err := b.UserService.GetAllStudents()
	if err != nil {
		b.sendError(chatID, "❌ Ошибка при получении списка учеников")
		return
	}

	if len(students) == 0 {
		msg := tgbotapi.NewMessage(chatID, "📭 Список учеников пуст")
		b.api.Send(msg)
		return
	}

	// Сохраняем учеников в сессии
	session := b.getOrCreateSession(chatID)
	session.State = StateSelectingStudent
	session.StudentsForSelection = students

	// Показываем список учеников
	b.showStudentsForSelection(chatID, students)
}
func (b *Bot) showStudentsForSelection(chatID int64, students []*models.User) {
	// Формируем сообщение со списком учеников
	msgText := "👥 *Выберите ученика:*\n\n"
	for i, student := range students {
		displayName := getStudentDisplayName(student)
		msgText += fmt.Sprintf("%d. %s\n", i+1, displayName)
	}
	msgText += "\nВведите номер ученика или отправьте '❌ Отмена'"

	msg := tgbotapi.NewMessage(chatID, msgText)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = createCancelKeyboard()

	b.api.Send(msg)
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

// func (b *Bot) handleAddSubscription(chatID int64) {
// 	user, _, _, _, err := b.UserService.GetUserProfile(chatID)
// 	if err != nil {
// 		b.sendError(chatID, "❌ Ошибка проверки прав доступа")
// 		return
// 	}

// 	if user.Role != "coach" {
// 		b.sendError(chatID, "❌ Эта функция доступна только тренерам")
// 		return
// 	}
// 	session := b.getOrCreateSession(chatID)
// 	session.State = StateSelectingStudent

// 	students, err := b.UserService.GetAllStudents()
// 	if err != nil {
// 		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка при получении списка учеников")
// 		b.api.Send(msg)
// 		b.resetSession(chatID)
// 		return
// 	}

// 	if len(students) == 0 {
// 		msg := tgbotapi.NewMessage(chatID, "📝 Нет доступных учеников")
// 		b.api.Send(msg)
// 		b.resetSession(chatID)
// 		return
// 	}

// 	// Создаем клавиатуру со списком учеников
// 	msg := tgbotapi.NewMessage(chatID, "👥 Выберите ученика:")
// 	keyboard := b.createStudentsKeyboard(students)
// 	msg.ReplyMarkup = keyboard
// 	b.api.Send(msg)
// }

// ///обработка выбора ученика
// func (b *Bot) handleStudentSelection(chatID int64, messageText string) {
// 	session := b.getOrCreateSession(chatID)
// 	if session.State != StateSelectingStudent {
// 		return
// 	}

// 	if messageText == "❌ Отмена" {
// 		b.cancelOperation(chatID)
// 		return
// 	}

// 	// Ищем ученика по имени (упрощенная логика)
// 	students, err := b.UserService.GetAllStudents()
// 	if err != nil {
// 		b.sendError(chatID, "Ошибка при поиске ученика")
// 		return
// 	}

// 	var selectedStudent *models.User
// 	for _, student := range students {
// 		expectedText := fmt.Sprintf("👤 %s %s", student.FirstName, student.LastName)
// 		if messageText == expectedText {
// 			selectedStudent = student
// 			break
// 		}
// 	}

// 	if selectedStudent == nil {
// 		msg := tgbotapi.NewMessage(chatID, "❌ Ученик не найден")
// 		b.api.Send(msg)
// 		return
// 	}
// 	session.SelectedStudentID = selectedStudent.ID
// 	session.State = StateSelectingSubscriptionType

// 	// Показываем выбор типа абонемента
// 	b.showSubscriptionTypes(chatID, selectedStudent)
// }

func (b *Bot) handleStudentSelection(chatID int64, messageText string) {
	session := b.getOrCreateSession(chatID)
	if session.State != StateSelectingStudent {
		return
	}

	if messageText == "❌ Отмена" {
		b.cancelOperation(chatID)
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
	)

	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}

func (b *Bot) handleSubscriptionTypeSelection(chatID int64, messageText string) {
	session := b.getOrCreateSession(chatID)
	if session.State != StateSelectingSubscriptionType {
		return
	}

	if messageText == "❌ Отмена" {
		b.cancelOperation(chatID)
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
	b.api.Send(msg)
}

func (b *Bot) handleConfirmation(chatID int64, messageText string) {
	session := b.getOrCreateSession(chatID)
	if session.State != StateConfirming {
		return
	}

	switch messageText {
	case "✅ Подтвердить":
		b.addSubscription(chatID, session)
	case "❌ Отмена":
		b.cancelOperation(chatID)
	default:
		b.sendError(chatID, "❌ Неизвестная команда")
	}
}

func (b *Bot) addSubscription(chatID int64, session *UserSession) {
	// var err error
	studentFromStudents, err := b.StudentService.GetStudentByUserID(session.SelectedStudentID)

	switch session.SelectedSubscriptionType {
	case "1":
		err = b.SubscriptionService.Create1For30Days(studentFromStudents.ID)
	case "12":
		err = b.SubscriptionService.Create12Unlimited(studentFromStudents.ID)
	case "16":
		err = b.SubscriptionService.Create16For30Days(studentFromStudents.ID)
	}

	if err != nil {
		b.sendError(chatID, "❌ Ошибка при добавлении абонемента: "+err.Error())
	} else {
		msg := tgbotapi.NewMessage(chatID, "✅ Абонемент успешно добавлен!")
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		b.api.Send(msg)
	}

	b.resetSession(chatID)
}

/////////////////////

func (b *Bot) resetSession(chatID int64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.userSessions, chatID)
}

func (b *Bot) cancelOperation(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "❌ Операция отменена")
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	b.api.Send(msg)
	b.resetSession(chatID)
}

func (b *Bot) sendError(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	b.api.Send(msg)
}
