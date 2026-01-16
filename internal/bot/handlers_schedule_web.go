package bot

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) handleCalendarCommand(message *tgbotapi.Message) {
	chatID := message.Chat.ID
	userID := message.From.ID

	_, err := b.UserService.GetByTelegramID(int64(userID))
	if err != nil {
		b.sendMessage(chatID, "Сначала зарегистрируйтесь в боте")
		return
	}

	// Формируем URL без user_id (будет использоваться initData из Telegram WebApp)
	url := fmt.Sprintf("%s/calendar", b.webBaseURL)

	/*
		// Старый код с простой отправкой URL
		text := url

		msg := tgbotapi.NewMessage(chatID, text)
		b.api.Send(msg)
	*/

	// Новый код с WebApp кнопкой
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(
				"📅 Открыть календарь",
				url,
			),
		),
	)

	msg := tgbotapi.NewMessage(chatID, "Нажмите кнопку ниже, чтобы открыть календарь тренировок:")
	msg.ReplyMarkup = keyboard

	// Добавляем fallback URL на случай, если WebApp не поддерживается
	msg.ParseMode = "HTML"
	msg.Text = "Нажмите кнопку ниже, чтобы открыть календарь тренировок:\n\n" +
		"<i>Если кнопка не работает, откройте ссылку в браузере:</i>\n" +
		fmt.Sprintf("<code>%s</code>", url)

	b.api.Send(msg)
}
