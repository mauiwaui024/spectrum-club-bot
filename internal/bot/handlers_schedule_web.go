package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// WebAppInfo представляет информацию о WebApp для кнопки
// Используется для создания WebApp кнопок, которые передают initData
type WebAppInfo struct {
	URL string `json:"url"`
}

// InlineKeyboardButtonWebApp представляет кнопку с WebApp для Telegram Bot API
// Библиотека go-telegram-bot-api/v5 не имеет встроенной поддержки web_app поля,
// поэтому создаем структуру с правильными JSON тегами для прямого использования
type InlineKeyboardButtonWebApp struct {
	Text   string      `json:"text"`
	WebApp *WebAppInfo `json:"web_app,omitempty"`
}

// InlineKeyboardMarkupWebApp представляет клавиатуру с WebApp кнопками
// Используется для отправки сообщений с WebApp кнопками через прямой API вызов
type InlineKeyboardMarkupWebApp struct {
	InlineKeyboard [][]InlineKeyboardButtonWebApp `json:"inline_keyboard"`
}

func (b *Bot) handleCalendarCommand(message *tgbotapi.Message) {
	chatID := message.Chat.ID
	userID := message.From.ID

	_, err := b.UserService.GetByTelegramID(int64(userID))
	if err != nil {
		b.sendMessage(chatID, "Сначала зарегистрируйтесь в боте")
		return
	}
	fmt.Println(userID)
	// Формируем URL без user_id (будет использоваться initData из Telegram WebApp)
	calendarURL := buildCalendarURL(b.webBaseURL)

	// Создаем WebApp кнопку для передачи initData
	// ВАЖНО: Для работы WebApp нужен HTTPS URL (не localhost)!
	// Telegram передает initData только для WebApp кнопок, не для обычных URL кнопок
	// Библиотека go-telegram-bot-api/v5 не поддерживает web_app поле напрямую,
	// поэтому используем прямой HTTP вызов к Telegram Bot API
	webAppMarkup := InlineKeyboardMarkupWebApp{
		InlineKeyboard: [][]InlineKeyboardButtonWebApp{
			{
				{
					Text: "📅 Открыть календарь",
					WebApp: &WebAppInfo{
						URL: calendarURL,
					},
				},
			},
		},
	}

	// Создаем структуру для отправки сообщения с WebApp кнопкой
	requestData := map[string]interface{}{
		"chat_id":      chatID,
		"text":         "Нажмите кнопку ниже, чтобы открыть календарь тренировок:\n\n<i>Если кнопка не работает, откройте ссылку в браузере:</i>\n<code>" + calendarURL + "</code>",
		"parse_mode":   "HTML",
		"reply_markup": webAppMarkup,
	}

	requestJSON, err := json.Marshal(requestData)
	if err != nil {
		// Fallback на обычную URL кнопку при ошибке
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL(
					"📅 Открыть календарь",
					calendarURL,
				),
			),
		)
		msg := tgbotapi.NewMessage(chatID, "Нажмите кнопку ниже, чтобы открыть календарь тренировок:")
		msg.ReplyMarkup = keyboard
		b.api.Send(msg)
		return
	}

	// Отправляем через прямой HTTP вызов к Telegram Bot API
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", b.api.Token)
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(requestJSON))
	if err != nil {
		// Fallback на обычную URL кнопку при ошибке сети
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL(
					"📅 Открыть календарь",
					calendarURL,
				),
			),
		)
		msg := tgbotapi.NewMessage(chatID, "Нажмите кнопку ниже, чтобы открыть календарь тренировок:")
		msg.ReplyMarkup = keyboard
		b.api.Send(msg)
		return
	}
	defer resp.Body.Close()

	// Проверяем ответ
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		// Логируем ошибку, но не паникуем - используем fallback
		fmt.Printf("Ошибка отправки WebApp кнопки: %s\n", string(body))
		// Fallback на обычную URL кнопку
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL(
					"📅 Открыть календарь",
					calendarURL,
				),
			),
		)
		msg := tgbotapi.NewMessage(chatID, "Нажмите кнопку ниже, чтобы открыть календарь тренировок:")
		msg.ReplyMarkup = keyboard
		b.api.Send(msg)
		return
	}
}

func buildCalendarURL(baseURL string) string {
	trimmed := strings.TrimSpace(baseURL)
	if trimmed == "" {
		return "/calendar"
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return strings.TrimRight(trimmed, "/") + "/calendar"
	}

	parsed.Path = strings.TrimRight(parsed.Path, "/")
	if parsed.Path == "/calendar" {
		return parsed.String()
	}
	parsed.Path = parsed.Path + "/calendar"

	return parsed.String()
}
