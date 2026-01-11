package bot

import (
	"fmt"
	"spectrum-club-bot/internal/models"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) handleDeleteSubscription(chatID int64, user *models.User) {
	if user.Role != "coach" {
		b.sendError(chatID, "❌ Эта функция доступна только тренерам")
		return
	}

	// Получаем всех студентов
	students, err := b.UserService.GetAllStudents()
	if err != nil {
		b.sendError(chatID, "❌ Ошибка при получении списка учеников")
		b.resetSession(chatID)
		return
	}

	if len(students) == 0 {
		b.sendError(chatID, "📝 Нет доступных учеников")
		b.resetSession(chatID)
		return
	}

	// Сохраняем учеников в сессии
	session := b.getOrCreateSession(chatID)
	session.State = StateSelectingStudentForDeletion
	session.StudentsForSelection = students

	// Показываем список учеников
	b.showStudentsForDeletion(chatID, students)
}

func (b *Bot) showStudentsForDeletion(chatID int64, students []*models.User) {
	// Формируем сообщение со списком учеников
	msgText := "👥 *Выберите ученика для удаления абонемента:*\n\n"
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

func (b *Bot) handleStudentSelectionForDeletion(chatID int64, messageText string) {
	session := b.getOrCreateSession(chatID)
	if session.State != StateSelectingStudentForDeletion {
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
	session.SelectedStudentForDeletion = selectedStudent
	session.State = StateSelectingSubscriptionForDeletion

	// Получаем студента из базы студентов
	student, err := b.StudentService.GetStudentByUserID(selectedStudent.ID)
	if err != nil || student == nil {
		b.sendError(chatID, "❌ Ошибка получения данных ученика")
		return
	}

	// Получаем все абонементы студента
	subscriptions, err := b.SubscriptionService.GetSubscriptionsByStudentID(student.ID)
	if err != nil {
		b.sendError(chatID, "❌ Ошибка получения абонементов ученика")
		return
	}

	if len(subscriptions) == 0 {
		msg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf("❌ У ученика %s %s нет абонементов",
				selectedStudent.FirstName, selectedStudent.LastName))
		b.api.Send(msg)
		b.resetSession(chatID)
		return
	}

	session.AvailableSubscriptions = subscriptions
	b.showSubscriptionsListForDeletion(chatID, selectedStudent, subscriptions)
}

func (b *Bot) showSubscriptionsListForDeletion(chatID int64, student *models.User, subscriptions []*models.Subscription) {
	msgText := fmt.Sprintf("🎫 *Выберите абонемент для удаления у %s %s:*\n\n", student.FirstName, student.LastName)

	for i, subscription := range subscriptions {
		var status string
		if subscription.RemainingLessons <= 0 {
			status = "❌"
		} else if time.Now().After(subscription.EndDate) {
			status = "⏰"
		} else {
			status = "✅"
		}

		msgText += fmt.Sprintf("%d. %s %d/%d занятий (до %s)\n",
			i+1,
			status,
			subscription.RemainingLessons,
			subscription.TotalLessons,
			subscription.EndDate.Format("02.01.2006"))
	}

	msgText += "\nВведите номер абонемента или отправьте '❌ Отмена'"

	msg := tgbotapi.NewMessage(chatID, msgText)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = createCancelKeyboard()
	b.api.Send(msg)
}

func (b *Bot) handleSubscriptionSelectionForDeletion(chatID int64, messageText string) {
	session := b.getOrCreateSession(chatID)
	if session.State != StateSelectingSubscriptionForDeletion {
		return
	}

	if messageText == "❌ Отмена" {
		b.cancelOperation(chatID)
		return
	}

	// Парсим номер абонемента
	index, err := strconv.Atoi(messageText)
	if err != nil || index < 1 || index > len(session.AvailableSubscriptions) {
		b.sendError(chatID, "❌ Пожалуйста, введите корректный номер абонемента")
		return
	}

	// Получаем выбранный абонемент по индексу
	selectedSubscription := session.AvailableSubscriptions[index-1]
	session.SelectedSubscriptionID = selectedSubscription.ID
	session.State = StateConfirmingSubscriptionDeletion

	b.showDeletionConfirmation(chatID, session.SelectedStudentForDeletion, selectedSubscription)
}

// Остальные функции остаются без изменений

func (b *Bot) showDeletionConfirmation(chatID int64, student *models.User, subscription *models.Subscription) {
	var status string
	if subscription.RemainingLessons <= 0 {
		status = "❌ ЗАВЕРШЕН"
	} else if time.Now().After(subscription.EndDate) {
		status = "⏰ ИСТЕК"
	} else {
		status = "✅ АКТИВЕН"
	}

	msgText := fmt.Sprintf(
		"🚨 *Подтвердите удаление абонемента:*\n\n"+
			"👤 *Ученик:* %s %s\n"+
			"🎫 *Абонемент:* %d/%d занятий\n"+
			"📅 *Действует до:* %s\n"+
			"🔰 *Статус:* %s\n\n"+
			"❌ *Это действие нельзя отменить!*",
		student.FirstName,
		student.LastName,
		subscription.RemainingLessons,
		subscription.TotalLessons,
		subscription.EndDate.Format("02.01.2006"),
		status,
	)

	msg := tgbotapi.NewMessage(chatID, msgText)
	msg.ParseMode = "Markdown"

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("✅ Удалить абонемент"),
			tgbotapi.NewKeyboardButton("❌ Отмена"),
		),
	)
	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}

func (b *Bot) handleSubscriptionDeletionConfirmation(chatID int64, messageText string) {
	session := b.getOrCreateSession(chatID)
	if session.State != StateConfirmingSubscriptionDeletion {
		return
	}

	switch messageText {
	case "✅ Удалить абонемент":
		b.deleteSubscription(chatID, session)
	case "❌ Отмена":
		b.cancelOperation(chatID)
	default:
		b.sendError(chatID, "❌ Неизвестная команда")
	}
}

func (b *Bot) deleteSubscription(chatID int64, session *UserSession) {
	err := b.SubscriptionService.DeleteSubscription(session.SelectedSubscriptionID)
	if err != nil {
		b.sendError(chatID, "❌ Ошибка при удалении абонемента: "+err.Error())
	} else {
		msg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf("✅ Абонемент успешно удален у ученика %s %s!",
				session.SelectedStudentForDeletion.FirstName,
				session.SelectedStudentForDeletion.LastName))
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		b.api.Send(msg)
	}

	b.resetSession(chatID)
}

// Вспомогательная функция для создания клавиатуры отмены
// func createCancelKeyboard() tgbotapi.ReplyKeyboardMarkup {
// 	return tgbotapi.NewReplyKeyboard(
// 		tgbotapi.NewKeyboardButtonRow(
// 			tgbotapi.NewKeyboardButton("❌ Отмена"),
// 		),
// 	)
// }
