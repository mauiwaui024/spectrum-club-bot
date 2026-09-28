package bot

import (
	"fmt"
	"log"
	"spectrum-club-bot/internal/models"
	"strings"
	"time"
	"unicode/utf8"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Обработка сообщения здесь
func (b *Bot) handleMessage(message *tgbotapi.Message) {
	log.Printf("[%s] %s", message.From.UserName, message.Text)

	user, _, _, _, err := b.UserService.GetUserProfile(int64(message.From.ID))
	if err != nil {
		log.Printf("Ошибка получения пользователя: %v", err)
	}

	chatID := message.Chat.ID

	session := b.getOrCreateSession(chatID)

	// Глобальные команды и кнопки меню сбрасывают залипший флоу,
	// иначе /start и «➕ Добавить абонемент» не показывают список.
	if message.IsCommand() {
		switch message.Command() {
		case "start", "coach", "student", "schedule":
			b.resetSession(chatID)
			session = b.getOrCreateSession(chatID)
		}
	}
	if session.State != StateDefault && isFlowInterruptText(message.Text) {
		b.resetSession(chatID)
		session = b.getOrCreateSession(chatID)
	}

	// Если у пользователя активная сессия - обрабатываем по состоянию
	if session.State != StateDefault {
		switch session.State {
		case StateSelectingStudent:
			b.handleStudentSelection(chatID, message.Text)
			return
		case StateSelectingSubscriptionType:
			b.handleSubscriptionTypeSelection(chatID, message.Text)
			return
		case StateConfirming:
			b.handleConfirmation(chatID, message.Text)
			return
		case StateSelectingGroup:
			b.handleGroupSelection(chatID, message.Text)
			return
		case StateSelectingDate:
			b.handleDateSelection(chatID, message.Text)
			return
		case StateSelectingTime:
			b.handleTimeSelection(chatID, message.Text)
			return
		case StateSelectingDuration:
			b.handleDurationSelection(chatID, message.Text)
			return
		case StateConfirmingTraining:
			b.handleTrainingConfirmation(chatID, message.Text)
			return
		// Новые состояния для удаления абонемента
		case StateSelectingStudentForDeletion:
			b.handleStudentSelectionForDeletion(chatID, message.Text)
			return
		case StateSelectingSubscriptionForDeletion:
			b.handleSubscriptionSelectionForDeletion(chatID, message.Text)
			return
		case StateConfirmingSubscriptionDeletion:
			b.handleSubscriptionDeletionConfirmation(chatID, message.Text)
			return
			// Также добавляем обработку новых состояний в switch session.State:
		case StateSelectingWeeksCount:
			b.handleWeeksCountSelection(message.Chat.ID, message.Text)
			return
		case StateConfirmingWeeklySchedule:
			b.handleWeeklyScheduleConfirmation(message.Chat.ID, user, message.Text)
			return

		case StateSelectingTrainingDateToEdit:
			b.handleDateSelectionForTrainingForEdit(chatID, message.Text)
			return

		case StateSelectingTrainingToEdit:
			b.handleTrainingSelectionForEdit(chatID, message.Text)
			return

		case StateSelectingFieldToEdit:
			b.handleFieldSelectionForEdit(chatID, message.Text)
			return

		case StateEditingTime:
			b.handleTimeEdit(chatID, message.Text)
			return

		case StateEditingPlace:
			b.handlePlaceEdit(chatID, message.Text)
			return

		case StateSelectingScheduleDate:
			b.handleScheduleDateInput(chatID, message.Text)
			return

		case StateSelectingSchedulePeriod:
			b.handleSchedulePeriodInput(chatID, message.Text)
			return

		case StateConfirmingDeletion:
			b.handleDeletionConfirmation(chatID, message.Text)
			return
		case StateSettingBrowserPassword:
			b.handleBrowserPasswordInput(chatID, user, message.Text)
			return
			// Состояния для записи на тренировку - удалены, функционал перенесен в WebApp
		}
	}

	// Обрабатываем команды (только если нет активной сессии)
	if message.IsCommand() {
		switch message.Command() {
		case "student":
			if user == nil {
				user, err = b.UserService.RegisterOrUpdate(
					int64(message.From.ID),
					message.From.FirstName,
					message.From.LastName,
					message.From.UserName,
					"student",
				)
				if err != nil {
					log.Printf("Ошибка регистрации пользователя: %v", err)
					return
				}
				b.handleNewStudentCommand(chatID, user)
			} else if user.Role == "student" {
				b.sendMessage(chatID, "Вы уже зарегистрированы как студент")
			} else if user.Role == "coach" {
				b.sendMessage(chatID, "Вы уже зарегистрированы как тренер")
			}
		case "start":
			b.handleStartCommand(message.Chat.ID, user)
		case "schedule":
			b.handleCalendarCommand(message)
		case "coach":
			//регистрируем как тренера
			if user == nil {
				user, err = b.UserService.RegisterOrUpdate(
					int64(message.From.ID),
					message.From.FirstName,
					message.From.LastName,
					message.From.UserName,
					"coach",
				)
				if err != nil {
					log.Printf("Ошибка регистрации пользователя: %v", err)
					return
				}
				b.handleCoachCommand(message.Chat.ID, user)
			} else if user.Role == "student" {
				b.sendMessage(chatID, "Вы уже зарегистрированы как студент")
			} else if user.Role == "coach" {
				b.sendMessage(chatID, "Вы уже зарегистрированы как тренер")
			}

		default:
			b.sendWelcomeMessage(message.Chat.ID, user)
		}
		return
	}

	// Обрабатываем текст сообщений (только если нет активной сессии)
	if isCancelText(message.Text) {
		b.cancelOperation(message.Chat.ID, user)
		return
	}

	switch message.Text {
	case "👤 Личный кабинет":
		// Для учеников перенаправляем в WebApp, для тренеров показываем личный кабинет
		if user.Role == "student" {
			b.handleCalendarCommand(message)
		} else {
			b.showPersonalAccount(message.Chat.ID, user)
		}
	case "👥 Мои ученики":
		b.showAllStudens(message.Chat.ID, user)
	case "📅 Календарь":
		b.handleCalendarCommand(message)
	case "🔐 Задать пароль":
		b.startBrowserPasswordFlow(chatID, user)
	case "📅 Управление расписанием":
		b.showScheduleManagementMenu(message.Chat.ID, user)
	case "💳 Управление абонементами":
		b.showSubscriptionManagementMenu(message.Chat.ID, user)
	case "📋 Создать из шаблонов":
		b.handleCreateFromTemplates(message.Chat.ID, user)
	case "◀️ Назад в главное меню":
		b.sendWelcomeMessage(message.Chat.ID, user)
	case "◀️ Назад к управлению расписанием":
		b.showScheduleManagementMenu(message.Chat.ID, user)
	// Обработка кнопок из подменю расписания
	case "➕ Добавить тренировку":
		b.handleAddTraining(message.Chat.ID, user)
	case "✏️ Редактировать тренировку":
		b.handleEditTraining(message.Chat.ID, user)
	case "📅 Мое расписание":
		b.handleMySchedule(message.Chat.ID, user)
	case "📅 На конкретную дату":
		b.handleScheduleTypeSelection(message.Chat.ID, message.Text)
	case "📆 На период":
		b.handleScheduleTypeSelection(message.Chat.ID, message.Text)
	case "➕ Добавить абонемент":
		b.handleAddSubscription(message.Chat.ID)
	case "🗑️ Удалить абонемент":
		b.handleDeleteSubscription(message.Chat.ID, user)
	case "👥 Список учеников с абонементами":
		b.showAllStudens(message.Chat.ID, user)

		// Для студентов - функционал перенесен в WebApp
		// Все операции доступны через команду /schedule или кнопку "📅 Календарь"

	case "◀️ Назад":
		b.sendWelcomeMessage(message.Chat.ID, user)
	default:
		b.sendWelcomeMessage(message.Chat.ID, user)
	}
}

func (b *Bot) showScheduleManagementMenu(chatID int64, user *models.User) {
	if user.Role != "coach" {
		b.sendError(chatID, "❌ Эта функция доступна только тренерам")
		return
	}

	msg := tgbotapi.NewMessage(chatID, "📅 *Управление расписанием*\n\nВыберите действие:")
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = createScheduleManagementKeyboard()
	b.api.Send(msg)
}

func (b *Bot) showSubscriptionManagementMenu(chatID int64, user *models.User) {
	if user.Role != "coach" {
		b.sendError(chatID, "❌ Эта функция доступна только тренерам")
		return
	}

	msg := tgbotapi.NewMessage(chatID, "💳 *Управление абонементами учеников*\n\nВыберите действие:")
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = createSubscriptionManagementKeyboard()
	b.api.Send(msg)
}

// ///
func (b *Bot) handleNewStudentCommand(chatID int64, user *models.User) {
	// При команде /start убеждаемся, что пользователь зарегистрирован как студент
	if user.Role != "student" {
		err := b.UserService.SetRole(user.TelegramID, "student")
		if err != nil {
			log.Printf("Ошибка установки роли студента: %v", err)
		} else {
			user.Role = "student"
		}
	}
	b.sendWelcomeMessage(chatID, user)
}

func (b *Bot) handleStartCommand(chatID int64, user *models.User) {
	text := "Введите правильную команду"
	b.sendMainMenuWithRefresh(chatID, user, text)
}

func (b *Bot) handleCoachCommand(chatID int64, user *models.User) {
	// При команде /coach регистрируем пользователя как тренера
	err := b.UserService.RegisterAsCoach(user.TelegramID, "Скалолазание", "Профессиональный тренер", "Опытный тренер по скалолазанию")
	if err != nil {
		log.Printf("Ошибка регистрации тренера: %v", err)
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка при регистрации тренера")
		b.api.Send(msg)
		return
	}
	user.Role = "coach"
	b.sendMainMenuWithRefresh(chatID, user, "✅ Теперь вы зарегистрированы как тренер!")
}

func (b *Bot) sendWelcomeMessage(chatID int64, user *models.User) {
	var text string
	if user == nil {
		b.sendError(chatID, "Не смогли получить юзера или неизвестная команда. Попробуйте /start")
		return
	}

	if user.Role == "coach" {
		text = `🏔 Добро пожаловать, тренер!

Выберите нужный раздел:`
	} else {
		text = `🏔 Добро пожаловать в клуб скалолазания!

📱 Весь функционал доступен через календарь.
Нажмите кнопку "📅 Календарь" или отправьте команду /schedule для открытия приложения.

В приложении вы можете:
• 📅 Просматривать расписание тренировок
• 📝 Записываться на тренировки
• 📋 Просматривать свои записи
• 🎫 Просматривать информацию об абонементе
• 👤 Просматривать свой профиль`
	}

	b.sendMainMenuWithRefresh(chatID, user, text)
}

func (b *Bot) sendMainMenuWithRefresh(chatID int64, user *models.User, text string) {
	// Force Telegram clients to drop stale keyboard cache first.
	remove := tgbotapi.NewMessage(chatID, " ")
	remove.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	b.api.Send(remove)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = createMainKeyboard(user.Role)
	b.api.Send(msg)
}

func (b *Bot) showPersonalAccount(chatID int64, user *models.User) {
	userProfile, _, _, _, err := b.UserService.GetUserProfile(user.TelegramID)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка при загрузке данных")
		b.api.Send(msg)
		return
	}

	var text string
	if userProfile.Role == "coach" {
		coach, err := b.CoachService.GetCoachByUserID(userProfile.ID)
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "❌ Ошибка при загрузке данных тренера")
			b.api.Send(msg)
			return
		}

		text = "👤 *Личный кабинет тренера*\n\n"
		text += "👤 *Имя:* " + userProfile.FirstName + " " + userProfile.LastName + "\n"

		if coach != nil {
			if coach.Specialty != "" {
				text += "🎯 *Специализация:* " + coach.Specialty + "\n"
			}
			if coach.Experience != "" {
				text += "📊 *Опыт:* " + coach.Experience + "\n"
			}
			if coach.Description != "" {
				text += "📝 *Описание:* " + coach.Description + "\n"
			}
		}
	} else {
		text = "👤 *Личный кабинет*\n\n"
		text += "👤 *Имя:* " + userProfile.FirstName + " " + userProfile.LastName + "\n"
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = createPersonalAccountKeyboard(userProfile.Role)
	b.api.Send(msg)
}

func (b *Bot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	b.sendTelegram(msg)
}

func (b *Bot) sendTelegram(c tgbotapi.Chattable) error {
	_, err := b.api.Send(c)
	if err != nil {
		log.Printf("[telegram Send] %v", err)
	}
	return err
}

const telegramMaxMessageLen = 3900

func (b *Bot) sendLongPlain(chatID int64, text string, markup interface{}) error {
	parts := splitTelegramText(text, telegramMaxMessageLen)
	for i, part := range parts {
		msg := tgbotapi.NewMessage(chatID, part)
		if i == len(parts)-1 && markup != nil {
			msg.ReplyMarkup = markup
		}
		if err := b.sendTelegram(msg); err != nil {
			return err
		}
	}
	return nil
}

func splitTelegramText(text string, max int) []string {
	if max < 1 {
		max = telegramMaxMessageLen
	}
	text = strings.TrimRight(text, "\n")
	if text == "" {
		return []string{" "}
	}
	var chunks []string
	for text != "" {
		if len(text) <= max {
			chunks = append(chunks, text)
			break
		}
		window := text[:max]
		cut := strings.LastIndex(window, "\n")
		if cut <= 0 {
			cut = max
			for cut > 0 && !utf8.RuneStart(text[cut]) {
				cut--
			}
			if cut == 0 {
				cut = max
			}
		}
		chunks = append(chunks, text[:cut])
		text = strings.TrimLeft(text[cut:], "\n")
	}
	return chunks
}

func isFlowInterruptText(text string) bool {
	switch text {
	case "➕ Добавить абонемент",
		"🗑️ Удалить абонемент",
		"👥 Список учеников с абонементами",
		"👥 Мои ученики",
		"💳 Управление абонементами",
		"📅 Управление расписанием",
		"📅 Календарь",
		"👤 Личный кабинет",
		"🔐 Задать пароль",
		"📋 Создать из шаблонов",
		"➕ Добавить тренировку",
		"✏️ Редактировать тренировку",
		"📅 Мое расписание",
		"◀️ Назад в главное меню",
		"◀️ Назад",
		"◀️ Назад к управлению расписанием":
		return true
	default:
		return false
	}
}

func (b *Bot) showAllStudens(chatID int64, user *models.User) {
	if user.Role != "coach" {
		msg := tgbotapi.NewMessage(chatID, "❌ Не имеете права, как вообще сюда попали")
		b.api.Send(msg)
		return
	}

	students, err := b.UserService.GetAllStudents()
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка при выполнении запроса")
		b.api.Send(msg)
		return
	}

	if len(students) == 0 {
		msg := tgbotapi.NewMessage(chatID, "📝 Список учеников пуст")
		b.api.Send(msg)
		return
	}

	// Получаем все абонементы
	allSubscriptions, err := b.SubscriptionService.GetAll()
	if err != nil {
		errStr := err.Error()
		msg := tgbotapi.NewMessage(chatID, errStr)
		b.api.Send(msg)
		return
	}

	// Создаем мапу для быстрого доступа к студентам по user_id
	studentMap := make(map[int64]*models.Student)
	for _, studentUser := range students {
		student, err := b.StudentService.GetStudentByUserID(studentUser.ID)
		if err == nil && student != nil {
			studentMap[studentUser.ID] = student
		}
	}

	// Создаем мапу абонементов по student_id (теперь это слайс абонементов)
	subscriptionMap := make(map[int64][]*models.Subscription)
	for _, subscription := range allSubscriptions {
		subscriptionMap[subscription.StudentID] = append(subscriptionMap[subscription.StudentID], subscription)
	}

	// Формируем сообщение со списком учеников и их абонементами.
	// Без Markdown: имена из Telegram ломают разбор (* _ [).
	var message string = "👥 Список всех учеников:\n\n"

	for i, studentUser := range students {
		displayName := getStudentDisplayName(studentUser)

		message += fmt.Sprintf("%d. %s\n", i+1, displayName)

		// Получаем информацию об абонементах
		student := studentMap[studentUser.ID]
		if student != nil {
			subscriptions := subscriptionMap[student.ID]

			// Фильтруем активные абонементы или те, где остались занятия
			var activeSubscriptions []*models.Subscription
			for _, sub := range subscriptions {
				if sub.RemainingLessons > 0 {
					activeSubscriptions = append(activeSubscriptions, sub)
				}
			}

			if len(activeSubscriptions) > 0 {
				for _, subscription := range activeSubscriptions {
					// Определяем статус абонемента
					status := "✅ Активен"
					if time.Now().After(subscription.EndDate) {
						status = "⏰ Истек (остались занятия)"
					}

					message += fmt.Sprintf("   🎫 %d/%d занятий - %s\n",
						subscription.RemainingLessons,
						subscription.TotalLessons,
						status)
					message += fmt.Sprintf("   📅 Действует до: %s\n",
						subscription.EndDate.Format("02.01.2006"))
				}
			} else {
				// Проверяем, есть ли вообще абонементы у студента
				if len(subscriptions) > 0 {
					message += "   ❌ Все абонементы использованы\n"
				} else {
					message += "   ❌ Нет абонементов\n"
				}
			}
		} else {
			message += "   ⚠️ Не найден в базе студентов\n"
		}

		message += "\n"
	}

	if err := b.sendLongPlain(chatID, message, nil); err != nil {
		b.sendError(chatID, "❌ Не удалось показать список учеников")
	}
}
