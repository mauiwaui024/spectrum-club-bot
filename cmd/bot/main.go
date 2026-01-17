package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"spectrum-club-bot/internal/bot"
	"spectrum-club-bot/internal/models/config"
	"spectrum-club-bot/internal/repository/attendance"
	"spectrum-club-bot/internal/repository/coach"
	"spectrum-club-bot/internal/repository/group"
	"spectrum-club-bot/internal/repository/schedule"
	"spectrum-club-bot/internal/repository/schedule_template"
	"spectrum-club-bot/internal/repository/student"
	"spectrum-club-bot/internal/repository/subscription"
	"spectrum-club-bot/internal/repository/user"
	attendance_service "spectrum-club-bot/internal/service/attendance"
	coach_service "spectrum-club-bot/internal/service/coach"
	group_serivce "spectrum-club-bot/internal/service/group"
	schedule_service "spectrum-club-bot/internal/service/schedule"
	student_service "spectrum-club-bot/internal/service/student"
	subscription_service "spectrum-club-bot/internal/service/subscription"
	user_service "spectrum-club-bot/internal/service/user"
	"spectrum-club-bot/internal/web"
	database "spectrum-club-bot/pkg"
	"syscall"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	// Загружаем конфигурацию
	if err := config.Load(); err != nil {
		log.Fatalf("❌ Ошибка загрузки конфигурации: %v", err)
	}

	cfg := config.AppConfig
	log.Printf("🚀 Запуск в окружении: %s", cfg.Environment)

	// Подключаемся к БД
	db, err := database.NewPostgres()
	if err != nil {
		log.Fatalf("❌ Ошибка подключения к БД: %v", err)
	}
	defer db.Close()

	// Инициализация репозиториев
	userRepo := user.NewUserRepository(db)
	studentRepo := student.NewStudentRepository(db)
	coachRepo := coach.NewCoachRepository(db)
	subscriptionRepo := subscription.NewSubscriptionRepository(db)
	attendanceRepo := attendance.NewAttendanceRepository(db)
	scheduleRepo := schedule.NewTrainingScheduleRepository(db)
	trainingGroupRepo := group.NewTrainingGroupRepository(db)
	templateScheduleRepos := schedule_template.NewWeekScheduleRepository(db)
	// Инициализация сервисов
	userService := user_service.NewUserService(userRepo, studentRepo, coachRepo, subscriptionRepo)
	studentService := student_service.NewStudentService(studentRepo)
	coachService := coach_service.NewCoachService(coachRepo)
	subscriptionService := subscription_service.NewSubscriptionService(subscriptionRepo)
	trainingGroupService := group_serivce.NewTrainingGroupService(trainingGroupRepo)
	//new
	attendanceService := attendance_service.NewAttendanceService(attendanceRepo, scheduleRepo, subscriptionService)
	scheduleService := schedule_service.NewScheduleService(scheduleRepo, attendanceRepo, templateScheduleRepos, trainingGroupRepo)
	// Создаем веб-хендлер с botToken для проверки Telegram WebApp initData
	calendarHandler := web.NewHandler(
		scheduleService,
		coachService,
		attendanceService,
		studentService,
		userService,
		subscriptionService,
		cfg.Bot.Token,
	)

	telegramBot, err := bot.NewBot(
		userService,
		coachService,
		studentService,
		subscriptionService,
		attendanceService,
		scheduleService,
		trainingGroupService,
	)
	if err != nil {
		log.Fatal("❌ Failed to create bot:", err)
	}

	// Настраиваем HTTP сервер
	mux := http.NewServeMux()

	// API endpoints (должны быть перед статикой)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	// Debug endpoint для проверки initData (только в development)
	// Всегда доступен для локальной отладки
	mux.HandleFunc("/api/debug/initdata", func(w http.ResponseWriter, r *http.Request) {
		initData := r.Header.Get("X-Telegram-Init-Data")

		// Логирование для отладки
		log.Printf("[DEBUG /api/debug/initdata] Запрос получен")
		log.Printf("[DEBUG /api/debug/initdata] initData в заголовке: %v (длина: %d)",
			initData != "", len(initData))
		if initData != "" {
			log.Printf("[DEBUG /api/debug/initdata] initData (первые 100 символов): %s",
				func() string {
					if len(initData) > 100 {
						return initData[:100] + "..."
					}
					return initData
				}())

			// Проверка будет выполнена в verifyTelegramWebAppData (логи там)
			log.Printf("[DEBUG /api/debug/initdata] initData будет проверен в verifyTelegramWebAppData")
		} else {
			log.Printf("[DEBUG /api/debug/initdata] initData пустой - страница открыта не через Telegram")
		}

		// Получаем все заголовки для отладки
		allHeaders := make(map[string]string)
		for k, v := range r.Header {
			if len(v) > 0 {
				allHeaders[k] = v[0]
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*") // Для локальной разработки
		json.NewEncoder(w).Encode(map[string]interface{}{
			"initData":       initData,
			"initDataLength": len(initData),
			"hasInitData":    initData != "",
			"userAgent":      r.UserAgent(),
			"environment":    cfg.Environment,
			"headers":        allHeaders,
		})
	})

	// Тестовый endpoint для проверки initData вручную (POST с initData в теле)
	mux.HandleFunc("/api/test/initdata", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Получаем initData из заголовка или тела запроса
		initData := r.Header.Get("X-Telegram-Init-Data")
		if initData == "" {
			if err := r.ParseForm(); err == nil {
				initData = r.FormValue("initData")
			}
		}
		if initData == "" {
			// Пробуем получить из JSON body
			var body map[string]string
			if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
				initData = body["initData"]
			}
		}

		log.Printf("[TEST /api/test/initdata] Получен initData для тестирования, длина: %d", len(initData))

		if initData == "" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "initData не предоставлен",
				"usage": "Отправьте POST запрос с initData в заголовке X-Telegram-Init-Data или в теле запроса (form: initData или json: {\"initData\": \"...\"})",
			})
			return
		}

		log.Printf("[TEST /api/test/initdata] initData (первые 100 символов): %s",
			func() string {
				if len(initData) > 100 {
					return initData[:100] + "..."
				}
				return initData
			}())

		// Используем AuthAPI для проверки initData (он уже делает всю работу)
		// Создаем временный request с initData в заголовке
		testReq, _ := http.NewRequest("POST", "/api/auth", nil)
		testReq.Header.Set("X-Telegram-Init-Data", initData)

		// Вызываем AuthAPI напрямую - он проверит initData и вернет userID
		calendarHandler.AuthAPI(w, testReq)
	})
	mux.HandleFunc("/api/auth", calendarHandler.AuthAPI)
	mux.HandleFunc("/api/training/", calendarHandler.TrainingDetailsAPI)
	mux.HandleFunc("/api/calendar", calendarHandler.CalendarAPI)
	mux.HandleFunc("/api/check-registration", calendarHandler.CheckRegistration)
	mux.HandleFunc("/api/register", calendarHandler.RegisterForTraining)
	mux.HandleFunc("/api/cancel", calendarHandler.CancelRegistration)
	mux.HandleFunc("/api/mark-attendance", calendarHandler.MarkAttendanceAPI)

	// Статические файлы Angular (для production)
	// В development Angular dev server будет на порту 4200
	angularDir := http.Dir("frontend/dist/spectrum-club-calendar/browser")
	angularFS := http.FileServer(angularDir)

	// Для SPA: все запросы, кроме API, возвращают index.html
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, существует ли файл
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		file, err := angularDir.Open(path)
		if err != nil {
			// Файл не найден - возвращаем index.html для SPA роутинга
			indexFile, err := angularDir.Open("/index.html")
			if err != nil {
				http.NotFound(w, r)
				return
			}
			defer indexFile.Close()
			http.ServeContent(w, r, "index.html", time.Time{}, indexFile)
			return
		}
		defer file.Close()

		// Файл существует - отдаем его
		angularFS.ServeHTTP(w, r)
	})
	srv := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: mux,
	}

	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	go func() {
		log.Printf("🌐 HTTP сервер запущен на порту %s", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Ошибка запуска HTTP сервера: %v", err)
		}
	}()

	// Запускаем бота в горутине
	go func() {
		if err := telegramBot.Start(); err != nil {
			log.Printf("❌ Ошибка запуска бота: %v", err)
			stop()
		}
	}()

	// Ожидаем сигнал завершения
	<-ctx.Done()

	log.Println("🛑 Получен сигнал завершения...")

	// Даем время на завершение операций
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Graceful shutdown HTTP сервера
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("⚠️ Ошибка при остановке HTTP сервера: %v", err)
	}

	// Здесь можно добавить cleanup
	log.Println("👋 Корректное завершение работы")
}
