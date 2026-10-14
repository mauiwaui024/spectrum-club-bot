package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"spectrum-club-bot/internal/models/config"
	"spectrum-club-bot/internal/repository/attendance"
	"spectrum-club-bot/internal/repository/coach"
	"spectrum-club-bot/internal/repository/group"
	planrepo "spectrum-club-bot/internal/repository/plan"
	"spectrum-club-bot/internal/repository/schedule"
	"spectrum-club-bot/internal/repository/schedule_template"
	"spectrum-club-bot/internal/repository/student"
	"spectrum-club-bot/internal/repository/subscription"
	"spectrum-club-bot/internal/repository/user"
	attendance_service "spectrum-club-bot/internal/service/attendance"
	coach_service "spectrum-club-bot/internal/service/coach"
	group_serivce "spectrum-club-bot/internal/service/group"
	plan_service "spectrum-club-bot/internal/service/plan"
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
	if err := config.Load(); err != nil {
		log.Fatalf("❌ Ошибка загрузки конфигурации: %v", err)
	}

	cfg := config.AppConfig
	log.Printf("🚀 Запуск в окружении: %s", cfg.Environment)

	db, err := database.NewPostgres()
	if err != nil {
		log.Fatalf("❌ Ошибка подключения к БД: %v", err)
	}
	defer db.Close()

	userRepo := user.NewUserRepository(db)
	studentRepo := student.NewStudentRepository(db)
	coachRepo := coach.NewCoachRepository(db)
	subscriptionRepo := subscription.NewSubscriptionRepository(db)
	attendanceRepo := attendance.NewAttendanceRepository(db)
	scheduleRepo := schedule.NewTrainingScheduleRepository(db)
	trainingGroupRepo := group.NewTrainingGroupRepository(db)
	templateScheduleRepos := schedule_template.NewWeekScheduleRepository(db)

	userService := user_service.NewUserService(userRepo, studentRepo, coachRepo, subscriptionRepo)
	studentService := student_service.NewStudentService(studentRepo)
	coachService := coach_service.NewCoachService(coachRepo)
	subscriptionService := subscription_service.NewSubscriptionService(subscriptionRepo)
	trainingGroupService := group_serivce.NewTrainingGroupService(trainingGroupRepo)
	attendanceService := attendance_service.NewAttendanceService(attendanceRepo, scheduleRepo, subscriptionService)
	scheduleService := schedule_service.NewScheduleService(scheduleRepo, attendanceRepo, templateScheduleRepos, trainingGroupRepo)

	calendarHandler := web.NewHandler(
		scheduleService,
		coachService,
		attendanceService,
		studentService,
		userService,
		subscriptionService,
		trainingGroupService,
		cfg.BaseURL,
	)
	calendarHandler.SetPlanService(plan_service.NewPlanService(planrepo.NewRepository(db)))

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("/api/auth/status", calendarHandler.AuthStatusAPI)
	mux.HandleFunc("/api/pd-consent", calendarHandler.PDConsentAPI)
	mux.HandleFunc("/api/browser-auth/login", calendarHandler.BrowserLoginAPI)
	mux.HandleFunc("/api/browser-auth/logout", calendarHandler.BrowserLogoutAPI)
	mux.HandleFunc("/api/browser-auth/set-password", calendarHandler.BrowserSetPasswordAPI)
	mux.HandleFunc("/api/training/", calendarHandler.TrainingDetailsAPI)
	mux.HandleFunc("/api/calendar", calendarHandler.CalendarAPI)
	mux.HandleFunc("/api/check-registration", calendarHandler.CheckRegistration)
	mux.HandleFunc("/api/register", calendarHandler.RegisterForTraining)
	mux.HandleFunc("/api/cancel", calendarHandler.CancelRegistration)
	mux.HandleFunc("/api/mark-attendance", calendarHandler.MarkAttendanceAPI)
	mux.HandleFunc("/api/unmark-attendance", calendarHandler.UnmarkAttendanceAPI)
	mux.HandleFunc("/api/my-registrations", calendarHandler.MyRegistrationsAPI)
	mux.HandleFunc("/api/my-subscription", calendarHandler.MySubscriptionAPI)
	mux.HandleFunc("/api/my-profile", calendarHandler.MyProfileAPI)
	mux.HandleFunc("/api/update-profile", calendarHandler.UpdateProfileAPI)
	mux.HandleFunc("/api/students-subscriptions", calendarHandler.AllStudentsSubscriptionsAPI)
	mux.HandleFunc("/api/subscription/add-lessons", calendarHandler.AddLessonsAPI)
	mux.HandleFunc("/api/subscription/remove-lessons", calendarHandler.RemoveLessonsAPI)
	mux.HandleFunc("/api/subscription/end-date", calendarHandler.SetSubscriptionEndDateAPI)
	mux.HandleFunc("/api/plan-templates", calendarHandler.PlanTemplatesAPI)
	mux.HandleFunc("/api/plan-templates/", calendarHandler.PlanTemplateByIDAPI)
	mux.HandleFunc("/api/exercises", calendarHandler.ExercisesAPI)
	mux.HandleFunc("/api/exercises/", calendarHandler.ExerciseByIDAPI)
	mux.HandleFunc("/api/climbers", calendarHandler.ClimbersAPI)
	mux.HandleFunc("/api/climbers/", calendarHandler.ClimberByIDAPI)
	mux.HandleFunc("/api/groups", calendarHandler.GroupsAPI)
	mux.HandleFunc("/api/groups/", calendarHandler.GroupByIDAPI)
	mux.HandleFunc("/api/coach/schedule", calendarHandler.CoachScheduleAPI)
	mux.HandleFunc("/api/trainings", calendarHandler.TrainingsAPI)
	mux.HandleFunc("/api/trainings/", calendarHandler.TrainingByIDAPI)
	mux.HandleFunc("/api/templates", calendarHandler.TemplatesAPI)
	mux.HandleFunc("/api/templates/", calendarHandler.TemplateByIDAPI)
	mux.HandleFunc("/api/students", calendarHandler.StudentsAPI)
	mux.HandleFunc("/api/users/without-password", calendarHandler.UsersWithoutPasswordAPI)
	mux.HandleFunc("/api/users/", calendarHandler.UserByIDAPI)
	mux.HandleFunc("/api/password-links", calendarHandler.PasswordLinksAPI)
	mux.HandleFunc("/api/password-links/validate", calendarHandler.ValidatePasswordLinkAPI)
	mux.HandleFunc("/api/password-links/accept", calendarHandler.AcceptPasswordLinkAPI)
	mux.HandleFunc("/api/subscriptions", calendarHandler.SubscriptionsAPI)
	mux.HandleFunc("/api/subscriptions/", calendarHandler.SubscriptionByIDAPI)
	mux.HandleFunc("/api/invites", calendarHandler.CreateInviteAPI)
	mux.HandleFunc("/api/invites/validate", calendarHandler.ValidateInviteAPI)
	mux.HandleFunc("/api/register-invite", calendarHandler.RegisterInviteAPI)
	mux.HandleFunc("/api/analytics/summary", calendarHandler.AnalyticsSummaryAPI)
	mux.HandleFunc("/api/analytics/attendance.csv", calendarHandler.AnalyticsAttendanceCSVAPI)
	mux.HandleFunc("/api/analytics/inactive-students", calendarHandler.AnalyticsInactiveStudentsAPI)

	angularDir := http.Dir("frontend/dist/spectrum-club-calendar/browser")
	angularFS := http.FileServer(angularDir)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		file, err := angularDir.Open(path)
		if err != nil {
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

	<-ctx.Done()
	log.Println("🛑 Получен сигнал завершения...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("⚠️ Ошибка при остановке HTTP сервера: %v", err)
	}

	log.Println("👋 Корректное завершение работы")
}
