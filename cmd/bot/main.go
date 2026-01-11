package main

import (
	"context"
	"log"
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

	// Настраиваем graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

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
