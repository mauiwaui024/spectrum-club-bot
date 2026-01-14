package bot

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) handleCalendarCommand(message *tgbotapi.Message) {
	chatID := message.Chat.ID
	userID := message.From.ID

	user, err := b.UserService.GetByTelegramID(int64(userID))
	if err != nil {
		b.sendMessage(chatID, "Сначала зарегистрируйтесь в боте")
		return
	}

	// Формируем URL
	url := fmt.Sprintf("%s/calendar?user_id=%d", b.webBaseURL, user.ID)

	text := "Ваше расписание:\n\n" + url

	msg := tgbotapi.NewMessage(chatID, text)
	b.api.Send(msg)
}
