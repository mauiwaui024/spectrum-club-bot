package main

import (
	"context"
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
	attendanceService := attendance_service.NewAttendanceService(attendanceRepo, scheduleRepo)
	scheduleService := schedule_service.NewScheduleService(scheduleRepo, attendanceRepo, templateScheduleRepos, trainingGroupRepo)
	// Создаем веб-хендлер
	calendarHandler := web.NewHandler(
		scheduleService,
		coachService,
		attendanceService,
		studentService,
		userService,
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
	mux.HandleFunc("/api/training/", calendarHandler.TrainingDetailsAPI)
	mux.HandleFunc("/api/calendar", calendarHandler.CalendarAPI)
	mux.HandleFunc("/api/check-registration", calendarHandler.CheckRegistration)
	mux.HandleFunc("/api/register", calendarHandler.RegisterForTraining)
	mux.HandleFunc("/api/cancel", calendarHandler.CancelRegistration)

	// Статические файлы Angular (для production)
	// В development Angular dev server будет на порту 4200
	angularDir := http.Dir("frontend/dist/spectrum-club-calendar/browser")
	angularFS := http.FileServer(angularDir)
	
	// Раздаем статические файлы Angular
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
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Здесь можно добавить cleanup
	log.Println("👋 Корректное завершение работы")
}
