package bot

import (
	"fmt"
	"log"
	"spectrum-club-bot/internal/models/config"
	"spectrum-club-bot/internal/service"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Bot struct {
	api                 *tgbotapi.BotAPI
	UserService         service.UserService
	CoachService        service.CoachService
	StudentService      service.StudentService
	SubscriptionService service.SubscriptionService
	//
	AttendanceService    service.AttendanceService
	ScheduleService      service.TrainingScheduleService
	TrainingGroupService service.TrainingGroupService
	////
	userSessions map[int64]*UserSession // chatID -> session
	mu           sync.RWMutex

	webBaseURL string // Добавляем базовый URL для веб-сервера
}

func NewBot(
	userService service.UserService,
	coachService service.CoachService,
	studentService service.StudentService,
	subscriptionService service.SubscriptionService,
	attendanceService service.AttendanceService,
	scheduleService service.TrainingScheduleService,
	trainingGroupService service.TrainingGroupService,
) (*Bot, error) {
	cfg := config.AppConfig.Bot

	if cfg.Token == "" {
		log.Panic("BOT_TOKEN не установлен в конфигурации")
	}

	api, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot API: %w", err)
	}

	api.Debug = cfg.Debug
	// Определяем базовый URL для веб-сервера
	webBaseURL := "http://localhost:8080"
	if config.AppConfig.Environment == "production" {
		if cfg.BaseURL == "" {
			return nil, fmt.Errorf("Пустая ссылка для webview")
		}
		webBaseURL = cfg.BaseURL // Укажите ваш домен
		
		// ВАЖНО: Для Telegram WebApp нужен HTTPS URL!
		// Если URL не начинается с http:// или https://, добавляем https://
		if !strings.HasPrefix(webBaseURL, "http://") && !strings.HasPrefix(webBaseURL, "https://") {
			webBaseURL = "https://" + webBaseURL
			log.Printf("⚠️  URL не содержал протокол, добавлен https://")
		}
		
		// Проверяем, что используется HTTPS (Telegram WebApp требует HTTPS для передачи initData)
		if strings.HasPrefix(webBaseURL, "http://") && !strings.Contains(webBaseURL, "localhost") {
			log.Printf("⚠️  ВНИМАНИЕ: Используется HTTP вместо HTTPS! Telegram WebApp может не передавать initData для HTTP URL.")
		}
	}
	log.Printf("🤖 URL календаря : %s", webBaseURL)
	log.Printf("🤖 Бот инициализирован: %s (debug: %v)", api.Self.UserName, cfg.Debug)
	log.Printf("👑 Администраторы: %v", cfg.AdminIDs)

	return &Bot{
		api:                  api,
		UserService:          userService,
		CoachService:         coachService,
		StudentService:       studentService,
		userSessions:         make(map[int64]*UserSession),
		SubscriptionService:  subscriptionService,
		AttendanceService:    attendanceService,
		ScheduleService:      scheduleService,
		TrainingGroupService: trainingGroupService,
		webBaseURL:           webBaseURL,
	}, nil
}
func (b *Bot) Start() error {
	log.Printf("Авторизован как %s", b.api.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := b.api.GetUpdatesChan(u)
	if err != nil {
		return err
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		go b.handleMessage(update.Message)
	}

	return nil
}
