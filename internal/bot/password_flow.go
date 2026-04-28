package bot

import (
	"regexp"
	"spectrum-club-bot/internal/models"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	hasUpperRegexp = regexp.MustCompile(`[A-Z]`)
	hasLowerRegexp = regexp.MustCompile(`[a-z]`)
	hasDigitRegexp = regexp.MustCompile(`[0-9]`)
	hasSpecialRegexp = regexp.MustCompile(`[^a-zA-Z0-9]`)
)

func (b *Bot) startBrowserPasswordFlow(chatID int64, user *models.User) {
	if user == nil {
		b.sendError(chatID, "❌ Сначала выполните /start")
		return
	}

	session := b.getOrCreateSession(chatID)
	session.State = StateSettingBrowserPassword

	msg := tgbotapi.NewMessage(chatID,
		"🔐 *Задать пароль для входа в браузере*\n\n"+
			"Введите новый пароль.\n\n"+
			"Требования:\n"+
			"• от 7 до 19 символов\n"+
			"• хотя бы 1 буква в верхнем регистре\n"+
			"• хотя бы 1 буква в нижнем регистре\n"+
			"• хотя бы 1 цифра\n"+
			"• хотя бы 1 спецсимвол\n\n"+
			"Для отмены отправьте: ❌ Отмена")
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = createCancelKeyboard()
	b.api.Send(msg)
}

func (b *Bot) handleBrowserPasswordInput(chatID int64, user *models.User, messageText string) {
	if messageText == "❌ Отмена" {
		b.cancelOperation(chatID, user)
		return
	}
	if user == nil {
		b.sendError(chatID, "❌ Пользователь не найден, выполните /start")
		b.resetSession(chatID)
		return
	}

	password := strings.TrimSpace(messageText)
	if errText := validateBrowserPassword(password); errText != "" {
		b.sendError(chatID, errText)
		return
	}

	if err := b.UserService.SetBrowserPassword(user.ID, password); err != nil {
		b.sendError(chatID, "❌ Не удалось сохранить пароль: "+err.Error())
		return
	}

	msg := tgbotapi.NewMessage(chatID,
		"✅ Пароль успешно установлен.\n\n"+
			"Теперь вы можете войти в обычном браузере через страницу /login используя ваш Telegram username и этот пароль.")
	b.api.Send(msg)
	b.resetSession(chatID)
	b.sendWelcomeMessage(chatID, user)
}

func validateBrowserPassword(password string) string {
	if len(password) < 7 || len(password) > 19 {
		return "❌ Пароль должен быть длиной от 7 до 19 символов."
	}
	if !hasUpperRegexp.MatchString(password) {
		return "❌ Пароль должен содержать хотя бы одну заглавную букву."
	}
	if !hasLowerRegexp.MatchString(password) {
		return "❌ Пароль должен содержать хотя бы одну строчную букву."
	}
	if !hasDigitRegexp.MatchString(password) {
		return "❌ Пароль должен содержать хотя бы одну цифру."
	}
	if !hasSpecialRegexp.MatchString(password) {
		return "❌ Пароль должен содержать хотя бы один спецсимвол."
	}
	return ""
}
