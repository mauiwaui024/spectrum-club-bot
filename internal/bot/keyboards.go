package bot

import (
	"fmt"
	"spectrum-club-bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func createMainKeyboard(role string) tgbotapi.ReplyKeyboardMarkup {
	if role == "coach" {
		return tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("📅 Управление расписанием"),
				tgbotapi.NewKeyboardButton("💳 Управление абонементами"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("👤 Личный кабинет"),
				tgbotapi.NewKeyboardButton("👥 Мои ученики"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("📅 Календарь"),
			),
		)
	}

	return createStudentMainKeyboard()
}

func createStudentMainKeyboard() tgbotapi.ReplyKeyboardMarkup {
	// Для учеников весь функционал доступен через WebApp (календарь)
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📅 Календарь"),
		),
	)
}

func createPersonalAccountKeyboard(role string) tgbotapi.ReplyKeyboardMarkup {
	if role == "coach" {
		// В личном кабинете только информация и кнопка назад
		return tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("◀️ Назад"),
			),
		)
	}

	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📝 Записаться на занятие"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📊 История посещений"),
			tgbotapi.NewKeyboardButton("◀️ Назад"),
		),
	)
}

// Эта функция уже у нас есть в коде, просто убедимся что она доступна
func (b *Bot) createGroupsKeyboard(groups []models.TrainingGroup) tgbotapi.ReplyKeyboardMarkup {
	var rows [][]tgbotapi.KeyboardButton

	for _, group := range groups {
		btn := tgbotapi.NewKeyboardButton(group.Name)
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(btn))
	}

	// Кнопка отмены
	cancelBtn := tgbotapi.NewKeyboardButton("❌ Отмена")
	rows = append(rows, tgbotapi.NewKeyboardButtonRow(cancelBtn))

	return tgbotapi.NewReplyKeyboard(rows...)
}

// func createScheduleActionsKeyboard() tgbotapi.ReplyKeyboardMarkup {
// 	return tgbotapi.NewReplyKeyboard(
// 		tgbotapi.NewKeyboardButtonRow(
// 			tgbotapi.NewKeyboardButton("📅 Расписание на 2 нед вперед"),
// 			tgbotapi.NewKeyboardButton("📅 Расписание на сегодня"),
// 		),
// 		tgbotapi.NewKeyboardButtonRow(
// 			// tgbotapi.NewKeyboardButton("📅 Расписание на неделю"),
// 			tgbotapi.NewKeyboardButton("◀️ Назад"),
// 		),
// 	)
// }

// Клавиатура для управления расписанием
func createScheduleManagementKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("➕ Добавить тренировку"),
			tgbotapi.NewKeyboardButton("✏️ Редактировать тренировку"), // Новая кнопка
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📅 Мое расписание"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📋 Создать из шаблонов"),
			tgbotapi.NewKeyboardButton("◀️ Назад в главное меню"),
		),
	)
}

// Клавиатура для управления абонементами
func createSubscriptionManagementKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("➕ Добавить абонемент"),
			tgbotapi.NewKeyboardButton("🗑️ Удалить абонемент"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("👥 Список учеников с абонементами"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("◀️ Назад в главное меню"),
		),
	)
}

// Клавиатура для студентов (для выбора в различных флоу)
func (b *Bot) createStudentsKeyboard(students []*models.User) tgbotapi.ReplyKeyboardMarkup {
	var rows [][]tgbotapi.KeyboardButton

	for _, student := range students {
		btn := tgbotapi.NewKeyboardButton(fmt.Sprintf("👤 %s %s", student.FirstName, student.LastName))
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(btn))
	}

	// Кнопка отмены
	cancelBtn := tgbotapi.NewKeyboardButton("❌ Отмена")
	rows = append(rows, tgbotapi.NewKeyboardButtonRow(cancelBtn))

	return tgbotapi.NewReplyKeyboard(rows...)
}

// ///////////////////////////////////////////

func createScheduleKeyboard(role string) tgbotapi.ReplyKeyboardMarkup {
	var rows [][]tgbotapi.KeyboardButton

	if role == "coach" {
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📅 Мое расписание"),
			tgbotapi.NewKeyboardButton("➕ Добавить тренировку"),
		))
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📋 Создать из шаблонов"),
			tgbotapi.NewKeyboardButton("✏️ Редактировать тренировку"),
		))
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("◀️ Назад в главное меню"),
		))
	} else {
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📅 Мое расписание"),
			tgbotapi.NewKeyboardButton("📝 Записаться на тренировку"),
		))
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("◀️ Назад в главное меню"),
		))
	}

	return tgbotapi.NewReplyKeyboard(rows...)
}
