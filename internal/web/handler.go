package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"spectrum-club-bot/internal/models"
	"spectrum-club-bot/internal/service"
)

var errNoSessionService = errors.New("session service not configured")

type Handler struct {
	scheduleService     service.TrainingScheduleService
	coachService        service.CoachService
	attendanceService   service.AttendanceService
	studentService      service.StudentService
	userService         service.UserService
	subscriptionService service.SubscriptionService
	groupService        service.TrainingGroupService
	planService         service.PlanService
	sessionService      service.SessionService
	inviteSecret        string
	baseURL             string
	allowLegacyUserID   bool
	usedInviteJTIs      map[string]time.Time
	usedInviteMu        sync.Mutex
}

func NewHandler(
	scheduleService service.TrainingScheduleService,
	coachService service.CoachService,
	attendanceService service.AttendanceService,
	studentService service.StudentService,
	userService service.UserService,
	subscriptionService service.SubscriptionService,
	groupService service.TrainingGroupService,
	sessionService service.SessionService,
	baseURL string,
) *Handler {
	inviteSecret := os.Getenv("INVITE_SECRET")
	if inviteSecret == "" {
		inviteSecret = "spectrum-invite-dev-secret"
	}
	env := strings.ToLower(os.Getenv("ENVIRONMENT"))
	return &Handler{
		scheduleService:     scheduleService,
		coachService:        coachService,
		attendanceService:   attendanceService,
		studentService:      studentService,
		userService:         userService,
		subscriptionService: subscriptionService,
		groupService:        groupService,
		sessionService:      sessionService,
		inviteSecret:        inviteSecret,
		baseURL:             strings.TrimRight(baseURL, "/"),
		allowLegacyUserID:   strings.ToLower(os.Getenv("ALLOW_LEGACY_USER_ID_FALLBACK")) == "true" || env != "production",
		usedInviteJTIs:      make(map[string]time.Time),
	}
}

// CalendarAPI возвращает данные календаря в формате JSON
func (h *Handler) CalendarAPI(w http.ResponseWriter, r *http.Request) {
	// Получаем userID из initData или query параметра (fallback)
	userID, userErr := h.getUserIDFromRequest(w, r)
	var userIDStr string
	if userErr == nil {
		userIDStr = strconv.FormatInt(userID, 10)
	}

	view := r.URL.Query().Get("view")
	dateStr := r.URL.Query().Get("date")

	if view == "" {
		view = "month"
	}

	var currentDate time.Time
	if dateStr == "" {
		currentDate = time.Now()
	} else {
		parsedDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			currentDate = time.Now()
		} else {
			currentDate = parsedDate
		}
	}

	// Определяем диапазон дат
	var startDate, endDate time.Time
	switch view {
	case "month":
		startDate = time.Date(currentDate.Year(), currentDate.Month(), 1, 0, 0, 0, 0, time.Local)
		endDate = startDate.AddDate(0, 1, 0)
	case "week":
		// Начинаем с понедельника
		weekday := int(currentDate.Weekday())
		if weekday == 0 {
			weekday = 7 // Воскресенье
		}
		startDate = currentDate.AddDate(0, 0, -(weekday - 1))
		endDate = startDate.AddDate(0, 0, 7)
	case "day":
		startDate = time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), 0, 0, 0, 0, time.Local)
		endDate = startDate.AddDate(0, 0, 1)
	default:
		// Schedule list view
		startDate = time.Now()
		endDate = startDate.AddDate(0, 1, 0) // Следующий месяц
	}

	// Получаем пользователя из initData или query параметра (fallback)
	var isCoach bool
	var userName string

	if userErr == nil {
		user, getUserErr := h.userService.GetByID(userID)
		if getUserErr == nil {
			isCoach = user.Role == "coach"
			userName = user.FirstName + " " + user.LastName
		}
	}

	// Получаем все тренировки для всех пользователей (тренеры и студенты видят все тренировки)
	trainings, err := h.scheduleService.GetTrainingsByDateRange(startDate, endDate)

	if err != nil {
		http.Error(w, "Ошибка получения расписания: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Подготавливаем JSON ответ
	response := h.prepareCalendarAPIResponse(view, currentDate, startDate, endDate, trainings, userIDStr, isCoach, userName)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Ошибка кодирования JSON: "+err.Error(), http.StatusInternalServerError)
	}
}

// Структуры для JSON API ответа
type CalendarAPIResponse struct {
	View         string              `json:"view"`
	CurrentDate  string              `json:"current_date"`
	StartDate    string              `json:"start_date"`
	EndDate      string              `json:"end_date"`
	PrevDate     string              `json:"prev_date"`
	NextDate     string              `json:"next_date"`
	IsCoach      bool                `json:"is_coach"`
	UserName     string              `json:"user_name"`
	UserID       string              `json:"user_id"`
	WeekDays     []WeekDayHeaderJSON `json:"week_days,omitempty"`
	CalendarDays []CalendarDayJSON   `json:"calendar_days,omitempty"`
	WeekDaysData []WeekDayDataJSON   `json:"week_days_data,omitempty"`
	Events       []CalendarEventJSON `json:"events,omitempty"`
	TimeSlots    []string            `json:"time_slots,omitempty"`
	TrainingDays []ScheduleDayJSON   `json:"training_days,omitempty"`
}

type WeekDayHeaderJSON struct {
	Name string `json:"name"`
	Day  string `json:"day"`
}

type CalendarDayJSON struct {
	Date         string              `json:"date"`
	IsToday      bool                `json:"is_today"`
	IsOtherMonth bool                `json:"is_other_month"`
	Events       []CalendarEventJSON `json:"events"`
}

type WeekDayDataJSON struct {
	Name    string              `json:"name"`
	Day     int                 `json:"day"`
	Date    string              `json:"date"`
	IsToday bool                `json:"is_today"`
	Events  []CalendarEventJSON `json:"events"`
}

type CalendarEventJSON struct {
	ID         int     `json:"id"`
	Title      string  `json:"title"`
	Time       string  `json:"time"`
	Coach      string  `json:"coach"`
	Top        float64 `json:"top,omitempty"`
	Height     float64 `json:"height,omitempty"`
	ColorIndex int     `json:"color_index"`
	UserID     string  `json:"user_id"`
}

type ScheduleDayJSON struct {
	Date      string             `json:"date"`
	Trainings []TrainingViewJSON `json:"trainings"`
}

type TrainingViewJSON struct {
	ID               int    `json:"id"`
	GroupName        string `json:"group_name"`
	StartTime        string `json:"start_time"`
	EndTime          string `json:"end_time"`
	CoachName        string `json:"coach_name"`
	Participants     int    `json:"participants"`
	ParticipantNames string `json:"participant_names"`
	MaxParticipants  int    `json:"max_participants"`
	CanRegister      bool   `json:"can_register"`
	IsRegistered     bool   `json:"is_registered"`
	IsFull           bool   `json:"is_full"`
	ColorIndex       int    `json:"color_index"`
}

func (h *Handler) prepareCalendarAPIResponse(view string, currentDate, startDate, endDate time.Time,
	trainings []models.TrainingSchedule, userIDStr string, isCoach bool, userName string) CalendarAPIResponse {

	response := CalendarAPIResponse{
		View:        view,
		CurrentDate: currentDate.Format("2006-01-02"),
		StartDate:   startDate.Format("2006-01-02"),
		EndDate:     endDate.Format("2006-01-02"),
		PrevDate:    h.getPrevDate(view, currentDate).Format("2006-01-02"),
		NextDate:    h.getNextDate(view, currentDate).Format("2006-01-02"),
		IsCoach:     isCoach,
		UserName:    userName,
		UserID:      userIDStr,
	}

	switch view {
	case "month":
		response.WeekDays = h.getWeekDayHeadersJSON()
		response.CalendarDays = h.prepareMonthViewJSON(currentDate, trainings, userIDStr)

	case "week":
		response.WeekDaysData = h.prepareWeekViewJSON(startDate, trainings, userIDStr)
		response.TimeSlots = h.generateTimeSlotsJSON()

	case "day":
		response.Events = h.prepareDayViewJSON(currentDate, trainings, userIDStr)
		response.TimeSlots = h.generateTimeSlotsJSON()

	default:
		// Для спискового вида (schedule)
		response.TrainingDays = h.prepareScheduleViewJSON(trainings, userIDStr, isCoach)
	}

	return response
}

func (h *Handler) getWeekDayHeadersJSON() []WeekDayHeaderJSON {
	return []WeekDayHeaderJSON{
		{Name: "ПН", Day: ""},
		{Name: "ВТ", Day: ""},
		{Name: "СР", Day: ""},
		{Name: "ЧТ", Day: ""},
		{Name: "ПТ", Day: ""},
		{Name: "СБ", Day: ""},
		{Name: "ВС", Day: ""},
	}
}

func (h *Handler) generateTimeSlotsJSON() []string {
	var slots []string
	for hour := 8; hour <= 22; hour++ {
		slots = append(slots, fmt.Sprintf("%02d:00", hour))
	}
	return slots
}

func (h *Handler) prepareMonthViewJSON(date time.Time, trainings []models.TrainingSchedule, userID string) []CalendarDayJSON {
	firstDay := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.Local)

	// Начинаем с понедельника недели, содержащей первый день
	weekday := int(firstDay.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	startDay := firstDay.AddDate(0, 0, -(weekday - 1))

	var days []CalendarDayJSON

	// 6 недель = 42 дня
	for i := 0; i < 42; i++ {
		day := startDay.AddDate(0, 0, i)

		dayData := CalendarDayJSON{
			Date:         day.Format("2006-01-02"),
			IsToday:      day.Format("2006-01-02") == time.Now().Format("2006-01-02"),
			IsOtherMonth: day.Month() != date.Month(),
		}

		var events []CalendarEventJSON
		for _, training := range trainings {
			trainingDate := training.TrainingDate.Format("2006-01-02")
			dayStr := day.Format("2006-01-02")

			if trainingDate == dayStr {
				events = append(events, CalendarEventJSON{
					ID:         training.ID,
					Title:      training.GroupName,
					Time:       fmt.Sprintf("%s - %s", training.StartTime.Format("15:04"), training.EndTime.Format("15:04")),
					Coach:      training.CoachName,
					ColorIndex: h.getColorIndex(training.GroupID),
					UserID:     userID,
				})
			}
		}

		dayData.Events = events
		days = append(days, dayData)
	}

	return days
}

func (h *Handler) prepareWeekViewJSON(startDate time.Time, trainings []models.TrainingSchedule, userID string) []WeekDayDataJSON {
	var weekDays []WeekDayDataJSON

	// 7 дней начиная с startDate (должен быть понедельник)
	for i := 0; i < 7; i++ {
		day := startDate.AddDate(0, 0, i)

		dayName := ""
		switch day.Weekday() {
		case time.Monday:
			dayName = "Пн"
		case time.Tuesday:
			dayName = "Вт"
		case time.Wednesday:
			dayName = "Ср"
		case time.Thursday:
			dayName = "Чт"
		case time.Friday:
			dayName = "Пт"
		case time.Saturday:
			dayName = "Сб"
		case time.Sunday:
			dayName = "Вс"
		}

		dayData := WeekDayDataJSON{
			Name:    dayName,
			Day:     day.Day(),
			Date:    day.Format("2006-01-02"),
			IsToday: day.Format("2006-01-02") == time.Now().Format("2006-01-02"),
		}

		// Добавляем тренировки для этого дня
		var events []CalendarEventJSON
		for _, training := range trainings {
			trainingDate := training.TrainingDate.Format("2006-01-02")
			dayStr := day.Format("2006-01-02")

			if trainingDate == dayStr {
				// Получаем только время (часы и минуты) из time.Time
				// PostgreSQL TIME сканируется как time.Time с нулевой датой
				// Не используем .In(time.Local) так как TIME не содержит информации о часовом поясе
				// Извлекаем часы и минуты напрямую
				startHour := training.StartTime.Hour()
				startMinute := training.StartTime.Minute()
				endHour := training.EndTime.Hour()
				endMinute := training.EndTime.Minute()

				// Рассчитываем позицию и высоту
				startMinutes := startHour*60 + startMinute
				endMinutes := endHour*60 + endMinute

				// Начало дня - 8:00 (480 минут), конец - 22:00 (1320 минут)
				startOfDay := 8 * 60
				endOfDay := 22 * 60

				// Ограничиваем время рамками дня
				if startMinutes < startOfDay {
					startMinutes = startOfDay
				}
				if endMinutes > endOfDay {
					endMinutes = endOfDay
				}
				if startMinutes >= endMinutes {
					startMinutes = startOfDay
					endMinutes = startOfDay + 60
				}

				// Позиция относительно 8:00 утра (480 минут)
				// 1px на минуту = 60px на час
				top := float64(startMinutes - startOfDay)
				height := float64(endMinutes - startMinutes)

				// Ограничиваем максимальную позицию (840px = 14 часов * 60px)
				if top > 840 {
					top = 840
				}
				if height > (840 - top) {
					height = 840 - top
				}
				if height < 30 {
					height = 30
				}

				events = append(events, CalendarEventJSON{
					ID:         training.ID,
					Title:      training.GroupName,
					Time:       fmt.Sprintf("%02d:%02d - %02d:%02d", startHour, startMinute, endHour, endMinute),
					Coach:      training.CoachName,
					Top:        top,
					Height:     height,
					ColorIndex: h.getColorIndex(training.GroupID),
					UserID:     userID,
				})
			}
		}

		dayData.Events = events
		weekDays = append(weekDays, dayData)
	}

	return weekDays
}

func (h *Handler) prepareDayViewJSON(date time.Time, trainings []models.TrainingSchedule, userID string) []CalendarEventJSON {
	var events []CalendarEventJSON

	dateStr := date.Format("2006-01-02")

	for _, training := range trainings {
		trainingDate := training.TrainingDate.Format("2006-01-02")

		if trainingDate == dateStr {
			// Получаем только время (часы и минуты) из time.Time
			// PostgreSQL TIME сканируется как time.Time с нулевой датой
			// Не используем .In(time.Local) так как TIME не содержит информации о часовом поясе
			// Извлекаем часы и минуты напрямую
			startHour := training.StartTime.Hour()
			startMinute := training.StartTime.Minute()
			endHour := training.EndTime.Hour()
			endMinute := training.EndTime.Minute()

			// Рассчитываем позицию и высоту
			startMinutes := startHour*60 + startMinute
			endMinutes := endHour*60 + endMinute

			// Начало дня - 8:00 (480 минут), конец - 22:00 (1320 минут)
			startOfDay := 8 * 60
			endOfDay := 22 * 60

			// Ограничиваем время рамками дня
			if startMinutes < startOfDay {
				startMinutes = startOfDay
			}
			if endMinutes > endOfDay {
				endMinutes = endOfDay
			}
			if startMinutes >= endMinutes {
				startMinutes = startOfDay
				endMinutes = startOfDay + 60
			}

			// Позиция относительно 8:00 утра (480 минут)
			// 1px на минуту = 60px на час
			top := float64(startMinutes - startOfDay)
			height := float64(endMinutes - startMinutes)

			// Ограничиваем максимальную позицию (840px = 14 часов * 60px)
			if top > 840 {
				top = 840
			}
			if height > (840 - top) {
				height = 840 - top
			}
			if height < 30 {
				height = 30
			}

			events = append(events, CalendarEventJSON{
				ID:         training.ID,
				Title:      training.GroupName,
				Time:       fmt.Sprintf("%02d:%02d - %02d:%02d", startHour, startMinute, endHour, endMinute),
				Coach:      training.CoachName,
				Top:        top,
				Height:     height,
				ColorIndex: h.getColorIndex(training.GroupID),
				UserID:     userID,
			})
		}
	}

	// Сортируем события по времени начала
	sort.Slice(events, func(i, j int) bool {
		// Парсим время из строки "HH:MM - HH:MM"
		timeI := events[i].Time
		timeJ := events[j].Time

		// Извлекаем время начала
		partsI := strings.Split(timeI, " - ")
		partsJ := strings.Split(timeJ, " - ")

		if len(partsI) > 0 && len(partsJ) > 0 {
			return partsI[0] < partsJ[0]
		}
		return false
	})

	return events
}

func (h *Handler) prepareScheduleViewJSON(trainings []models.TrainingSchedule, userID string, isCoach bool) []ScheduleDayJSON {
	// Группируем по дням
	grouped := make(map[string][]TrainingViewJSON)

	for _, training := range trainings {
		dateKey := training.TrainingDate.Format("2006-01-02")
		participants, _ := h.attendanceService.GetParticipants(training.ID)

		// Формируем строку с именами участников
		var participantNames string
		if len(participants) > 0 {
			var names []string
			for _, p := range participants {
				names = append(names, p.StudentName)
			}
			if len(names) > 3 {
				participantNames = strings.Join(names[:3], ", ") + " и ещё " + strconv.Itoa(len(names)-3)
			} else {
				participantNames = strings.Join(names, ", ")
			}
		}

		// Проверяем запись пользователя
		isRegistered := false
		canRegister := false

		if !isCoach && userID != "" {
			userIDInt, _ := strconv.ParseInt(userID, 10, 64)
			student, err := h.studentService.GetStudentByUserID(userIDInt)
			if err == nil {
				att, _ := h.attendanceService.GetStudentAttendanceForTraining(int(student.ID), training.ID)
				isRegistered = att != nil && att.Status == "registered"

				// Создаем полную дату и время начала тренировки для правильного сравнения
				trainingDateTime := time.Date(
					training.TrainingDate.Year(),
					training.TrainingDate.Month(),
					training.TrainingDate.Day(),
					training.StartTime.Hour(),
					training.StartTime.Minute(),
					training.StartTime.Second(),
					0,
					training.TrainingDate.Location(),
				)
				// Проверяем, можно ли записаться
				// Тренировка должна быть в будущем (дата и время начала)
				if !isRegistered && trainingDateTime.After(time.Now()) {
					canRegister = true
				}
			}
		}

		grouped[dateKey] = append(grouped[dateKey], TrainingViewJSON{
			ID:               training.ID,
			GroupName:        training.GroupName,
			StartTime:        training.StartTime.Format("15:04"),
			EndTime:          training.EndTime.Format("15:04"),
			CoachName:        training.CoachName,
			Participants:     len(participants),
			ParticipantNames: participantNames,
			CanRegister:      canRegister,
			IsRegistered:     isRegistered,
			IsFull:           false,
			ColorIndex:       h.getColorIndex(training.GroupID),
		})
	}

	// Преобразуем в массив и сортируем по дате
	var scheduleDays []ScheduleDayJSON
	for dateStr, trainings := range grouped {
		date, _ := time.Parse("2006-01-02", dateStr)

		// Форматируем дату
		formattedDate := date.Format("Monday, 2 January 2006")
		today := time.Now()
		if date.Format("2006-01-02") == today.Format("2006-01-02") {
			formattedDate = "Сегодня, " + date.Format("2 January")
		} else if date.Format("2006-01-02") == today.AddDate(0, 0, 1).Format("2006-01-02") {
			formattedDate = "Завтра, " + date.Format("2 January")
		}

		scheduleDays = append(scheduleDays, ScheduleDayJSON{
			Date:      formattedDate,
			Trainings: trainings,
		})
	}

	return scheduleDays
}

// prepareTemplateData method removed - теперь используется prepareCalendarAPIResponse для JSON API

func (h *Handler) getPrevDate(view string, date time.Time) time.Time {
	switch view {
	case "month":
		return date.AddDate(0, -1, 0)
	case "week":
		return date.AddDate(0, 0, -7)
	case "day":
		return date.AddDate(0, 0, -1)
	default:
		return date.AddDate(0, -1, 0)
	}
}

func (h *Handler) getNextDate(view string, date time.Time) time.Time {
	switch view {
	case "month":
		return date.AddDate(0, 1, 0)
	case "week":
		return date.AddDate(0, 0, 7)
	case "day":
		return date.AddDate(0, 0, 1)
	default:
		return date.AddDate(0, 1, 0)
	}
}

type WeekDayHeader struct {
	Name string
	Day  string
}

type WeekDayData struct {
	Name    string
	Day     int
	Date    time.Time
	IsToday bool
	Events  []CalendarEvent
}

type CalendarEvent struct {
	ID         int
	Title      string
	Time       string
	Coach      string
	Top        float64
	Height     float64
	ColorIndex int
	UserID     string
}

type CalendarDay struct {
	Date         time.Time
	IsToday      bool
	IsOtherMonth bool
	Events       []CalendarEvent
}

func (h *Handler) getColorIndex(groupID int) int {
	if groupID <= 0 {
		return 1
	}
	return (groupID % 5) + 1
}

// TrainingDetailsAPI возвращает детали тренировки в JSON
func (h *Handler) TrainingDetailsAPI(w http.ResponseWriter, r *http.Request) {
	// Получаем ID тренировки
	trainingIDStr := r.URL.Query().Get("training_id")
	if trainingIDStr == "" {
		// Попробуем получить из пути /api/training/{id}
		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) >= 3 {
			trainingIDStr = pathParts[3]
		}
	}

	trainingID, err := strconv.Atoi(trainingIDStr)
	if err != nil {
		http.Error(w, "Invalid training ID", http.StatusBadRequest)
		return
	}

	// Получаем userID из initData или query параметра (fallback)
	userID, err := h.getUserIDFromRequest(w, r)
	var userIDStr string
	var isCoach bool
	var isTrainingCoach bool
	var canMarkAttendance bool

	if err == nil {
		userIDStr = strconv.FormatInt(userID, 10)
		user, err := h.userService.GetByID(userID)
		if err == nil {
			isCoach = user.Role == "coach"
		}
	}
	// Получаем тренировку
	training, err := h.scheduleService.GetTrainingByID(trainingID)
	if err != nil {
		http.Error(w, "Training not found", http.StatusNotFound)
		return
	}

	// Получаем участников
	participants, err := h.attendanceService.GetParticipants(trainingID)
	if err != nil {
		participants = []models.AttendanceWithStudent{}
	}

	// Проверяем, записан ли пользователь
	isRegistered := false
	canRegister := false
	isFull := false

	// Создаем полную дату и время начала тренировки для правильного сравнения
	trainingDateTime := time.Date(
		training.TrainingDate.Year(),
		training.TrainingDate.Month(),
		training.TrainingDate.Day(),
		training.StartTime.Hour(),
		training.StartTime.Minute(),
		training.StartTime.Second(),
		0,
		training.TrainingDate.Location(),
	)

	now := time.Now()
	isPast := now.After(trainingDateTime)

	// Проверяем, является ли тренер тренером этой тренировки
	if isCoach && userIDStr != "" {
		coach, err := h.coachService.GetCoachByUserID(userID)
		if err == nil && training.CoachID != nil {
			isTrainingCoach = *training.CoachID == coach.ID
		}
	}

	// Проверяем, может ли тренер отмечать посещаемость
	// Условия: тренер, тренировка прошла, тренер является тренером этой тренировки
	canMarkAttendance = isCoach && isPast && isTrainingCoach

	// Проверяем регистрацию пользователя, если userID передан
	if userIDStr != "" && !isCoach {
		student, err := h.studentService.GetStudentByUserID(userID)
		if err == nil {
			att, _ := h.attendanceService.GetStudentAttendanceForTraining(int(student.ID), trainingID)
			isRegistered = att != nil && att.Status == "registered"
		}
	}

	// Проверяем, можно ли записаться (независимо от того, передан userID или нет)
	// userID нужен только для самой записи, но для определения can_register достаточно проверить:
	// 1. Пользователь - студент (не тренер)
	// 2. Тренировка в будущем (дата и время начала)
	// 3. Пользователь не записан (если userID передан, иначе считаем что не записан)
	// Примечание: проверка активного абонемента выполняется в RegisterForTraining,
	// чтобы пользователь видел кнопку и получал сообщение об ошибке при попытке записи
	if !isCoach && !isRegistered && trainingDateTime.After(now) {
		canRegister = true
	}

	// Формируем ответ
	response := map[string]interface{}{
		"training": map[string]interface{}{
			"id":            training.ID,
			"group_name":    training.GroupName,
			"training_date": training.TrainingDate.Format("2006-01-02"),
			"start_time":    training.StartTime.Format("15:04"),
			"end_time":      training.EndTime.Format("15:04"),
			"coach_name":    training.CoachName,
			"description":   training.Description,
		},
		"participants":        participants,
		"participants_count":  len(participants),
		"is_coach":            isCoach,
		"is_training_coach":   isTrainingCoach,
		"can_mark_attendance": canMarkAttendance,
		"is_registered":       isRegistered,
		"can_register":        canRegister,
		"is_full":             isFull,
		"is_past":             isPast,
		"current_time":        time.Now().Format(time.RFC3339),
		"plan":                h.trainingPlanOrNil(trainingID),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RegisterForTraining обрабатывает запись на тренировку
func (h *Handler) RegisterForTraining(w http.ResponseWriter, r *http.Request) {
	log.Printf("[RegisterForTraining] Запрос на регистрацию получен")
	log.Printf("[RegisterForTraining] Method: %s", r.Method)
	log.Printf("[RegisterForTraining] Headers: X-Telegram-Init-Data = %v (длина: %d)",
		r.Header.Get("X-Telegram-Init-Data") != "", len(r.Header.Get("X-Telegram-Init-Data")))

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Парсим multipart/form-data (для FormData из Angular)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		// Если не multipart, пробуем обычную форму
		if err := r.ParseForm(); err != nil {
			log.Printf("[RegisterForTraining] Ошибка парсинга формы: %v", err)
			http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Получаем параметры из формы
	trainingIDStr := r.FormValue("training_id")
	log.Printf("[RegisterForTraining] training_id из формы: %s", trainingIDStr)

	if trainingIDStr == "" {
		http.Error(w, "Missing training_id", http.StatusBadRequest)
		return
	}

	trainingID, err := strconv.Atoi(trainingIDStr)
	if err != nil {
		http.Error(w, "Invalid training ID", http.StatusBadRequest)
		return
	}

	// Получаем userID через unified auth (initData или browser session)
	userID, authErr := h.getUserIDFromRequest(w, r)
	if authErr != nil {
		log.Printf("[RegisterForTraining] auth failed: %v", authErr)
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Проверяем роль пользователя - только студенты могут записываться на тренировки
	user, err := h.userService.GetByID(userID)
	if err != nil {
		log.Printf("[RegisterForTraining] Ошибка получения пользователя: %v", err)
		http.Error(w, "User not found: "+err.Error(), http.StatusBadRequest)
		return
	}

	if user.Role == "coach" {
		log.Printf("[RegisterForTraining] ❌ Попытка записи тренера на тренировку (userID: %d, role: %s)", userID, user.Role)
		http.Error(w, "Только ученики могут записываться на тренировки. Тренеры не могут записываться на свои тренировки.", http.StatusForbidden)
		return
	}

	// Получаем студента по userID
	student, err := h.studentService.GetStudentByUserID(userID)
	if err != nil {
		http.Error(w, "Student not found: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Проверяем наличие активного абонемента с оставшимися занятиями
	activeSubscription, err := h.subscriptionService.GetActiveSubscription(student.ID)
	if err != nil || activeSubscription == nil {
		http.Error(w, "Пополните абонемент. Для записи на тренировку необходим активный абонемент с оставшимися занятиями.", http.StatusBadRequest)
		return
	}

	if activeSubscription.RemainingLessons <= -2 {
		http.Error(w, "Пополните абонемент. Для записи на тренировку необходим активный абонемент с оставшимися занятиями.", http.StatusBadRequest)
		return
	}

	// Проверяем, существует ли тренировка
	training, err := h.scheduleService.GetTrainingByID(trainingID)
	if err != nil {
		http.Error(w, "Training not found: "+err.Error(), http.StatusNotFound)
		return
	}

	if err := errIfTooLateForBooking(training, "Запись"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Проверяем, не записан ли уже студент
	existingAttendance, _ := h.attendanceService.GetStudentAttendanceForTraining(int(student.ID), trainingID)
	if existingAttendance != nil {
		if existingAttendance.Status == "registered" {
			http.Error(w, "Already registered for this training", http.StatusBadRequest)
			return
		}
	}

	// Записываем на тренировку
	err = h.attendanceService.SignUpForTraining(int(student.ID), trainingID)
	if err != nil {
		http.Error(w, "Failed to register: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully registered for training"))
}

// CancelRegistration обрабатывает отмену записи на тренировку
func (h *Handler) CancelRegistration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Парсим multipart/form-data (для FormData из Angular)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		// Если не multipart, пробуем обычную форму
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Получаем параметры из формы
	trainingIDStr := r.FormValue("training_id")
	if trainingIDStr == "" {
		http.Error(w, "Missing training_id", http.StatusBadRequest)
		return
	}

	trainingID, err := strconv.Atoi(trainingIDStr)
	if err != nil {
		http.Error(w, "Invalid training ID", http.StatusBadRequest)
		return
	}

	// Получаем userID через unified auth (initData или browser session)
	userID, authErr := h.getUserIDFromRequest(w, r)
	if authErr != nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Проверяем роль пользователя - только студенты могут отменять записи
	user, err := h.userService.GetByID(userID)
	if err != nil {
		http.Error(w, "User not found: "+err.Error(), http.StatusBadRequest)
		return
	}

	if user.Role == "coach" {
		log.Printf("[CancelRegistration] ❌ Попытка отмены записи тренером (userID: %d, role: %s)", userID, user.Role)
		http.Error(w, "Только ученики могут отменять записи на тренировки.", http.StatusForbidden)
		return
	}

	// Получаем студента по userID
	student, err := h.studentService.GetStudentByUserID(userID)
	if err != nil {
		http.Error(w, "Student not found: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Проверяем, записан ли студент
	attendance, err := h.attendanceService.GetStudentAttendanceForTraining(int(student.ID), trainingID)
	if err != nil || attendance == nil {
		http.Error(w, "Not registered for this training", http.StatusBadRequest)
		return
	}

	training, err := h.scheduleService.GetTrainingByID(trainingID)
	if err != nil {
		http.Error(w, "Training not found", http.StatusNotFound)
		return
	}
	if err := errIfTooLateForBooking(training, "Отмена"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Отменяем запись
	err = h.attendanceService.CancelSignUp(int(student.ID), trainingID)
	if err != nil {
		http.Error(w, "Failed to cancel: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Successfully cancelled registration"))
}

// MarkAttendanceAPI обрабатывает подтверждение посещаемости тренировки тренером
func (h *Handler) MarkAttendanceAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Парсим JSON body
	var requestData struct {
		TrainingID int   `json:"training_id"`
		StudentIDs []int `json:"student_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if requestData.TrainingID == 0 {
		http.Error(w, "Missing training_id", http.StatusBadRequest)
		return
	}

	// Получаем userID из initData (безопасный способ) или из query параметра (fallback)
	userID, authErr := h.getUserIDFromRequest(w, r)
	if authErr != nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Проверяем, что пользователь - тренер
	user, err := h.userService.GetByID(userID)
	if err != nil {
		http.Error(w, "User not found: "+err.Error(), http.StatusBadRequest)
		return
	}

	if user.Role != "coach" {
		http.Error(w, "Only coaches can mark attendance", http.StatusForbidden)
		return
	}

	// Получаем тренера по userID
	coach, err := h.coachService.GetCoachByUserID(userID)
	if err != nil {
		http.Error(w, "Coach not found: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Получаем тренировку
	training, err := h.scheduleService.GetTrainingByID(requestData.TrainingID)
	if err != nil {
		http.Error(w, "Training not found: "+err.Error(), http.StatusNotFound)
		return
	}

	// Проверяем, что тренер является тренером этой тренировки
	if training.CoachID == nil || *training.CoachID != coach.ID {
		http.Error(w, "Only the training coach can mark attendance", http.StatusForbidden)
		return
	}

	// Проверяем, что тренировка прошла
	trainingDateTime := time.Date(
		training.TrainingDate.Year(),
		training.TrainingDate.Month(),
		training.TrainingDate.Day(),
		training.StartTime.Hour(),
		training.StartTime.Minute(),
		training.StartTime.Second(),
		0,
		training.TrainingDate.Location(),
	)

	if trainingDateTime.After(time.Now()) {
		http.Error(w, "Cannot mark attendance for future training", http.StatusBadRequest)
		return
	}

	// Получаем всех записавшихся учеников
	participants, err := h.attendanceService.GetParticipants(requestData.TrainingID)
	if err != nil {
		http.Error(w, "Failed to get participants: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Создаем множество ID учеников, которые посетили
	attendedSet := make(map[int]bool)
	for _, studentID := range requestData.StudentIDs {
		attendedSet[studentID] = true
	}

	// Логирование для отладки
	log.Printf("[MarkAttendanceAPI] Получено student_ids для отметки: %v", requestData.StudentIDs)
	log.Printf("[MarkAttendanceAPI] Всего участников: %d", len(participants))

	// Отмечаем посещаемость только для выбранных учеников
	markedCount := 0
	failedStudents := []map[string]interface{}{}
	errors := []string{}

	for _, participant := range participants {
		// participant.StudentID это int из структуры Attendance
		studentID := participant.StudentID
		attended := attendedSet[studentID]

		log.Printf("[MarkAttendanceAPI] Участник ID=%d, studentID=%d, attended=%v (в attendedSet: %v)",
			participant.ID, studentID, attended, attendedSet[studentID])

		// Обновляем только выбранных учеников (attended = true)
		if !attended {
			log.Printf("[MarkAttendanceAPI] Ученик %d не выбран, пропускаем", studentID)
			continue
		}

		// recorded_by должен ссылаться на user_id, а не на coach.id
		// Преобразуем userID в int (userID уже получен из getUserIDFromRequest)
		recordedByUserID := int(userID)

		err := h.attendanceService.MarkAttendance(
			requestData.TrainingID,
			studentID,
			recordedByUserID,
			true, // Всегда true, так как мы обрабатываем только выбранных
			"",
		)
		if err != nil {
			log.Printf("[MarkAttendanceAPI] Ошибка отметки посещаемости для ученика %d: %v", studentID, err)
			failedStudents = append(failedStudents, map[string]interface{}{
				"student_id":   studentID,
				"student_name": participant.StudentName,
				"error":        err.Error(),
			})
			errors = append(errors, fmt.Sprintf("Ученик %s: %v", participant.StudentName, err))
			continue
		}
		markedCount++
		log.Printf("[MarkAttendanceAPI] Посещаемость успешно отмечена для ученика %d", studentID)

	}

	// Формируем ответ с информацией об ошибках
	response := map[string]interface{}{
		"success":         len(failedStudents) == 0,
		"marked_count":    markedCount,
		"total_count":     len(participants),
		"failed_count":    len(failedStudents),
		"failed_students": failedStudents,
	}

	if len(failedStudents) > 0 {
		response["errors"] = errors
		response["message"] = fmt.Sprintf("Посещаемость подтверждена для %d из %d учеников", markedCount, len(participants))
		log.Printf("[MarkAttendanceAPI] Частичный успех: отмечено %d, ошибок %d", markedCount, len(failedStudents))
	} else {
		response["message"] = "Посещаемость подтверждена"
		log.Printf("[MarkAttendanceAPI] Успешно отмечено %d учеников", markedCount)
	}

	// Устанавливаем HTTP статус
	w.Header().Set("Content-Type", "application/json")
	if len(failedStudents) > 0 {
		w.WriteHeader(http.StatusPartialContent) // 206 - частичный успех
	} else {
		w.WriteHeader(http.StatusOK)
	}
	json.NewEncoder(w).Encode(response)
}

// CheckRegistration проверяет статус записи студента на тренировку
func (h *Handler) CheckRegistration(w http.ResponseWriter, r *http.Request) {
	trainingIDStr := r.URL.Query().Get("training_id")
	if trainingIDStr == "" {
		http.Error(w, "Missing training_id", http.StatusBadRequest)
		return
	}

	trainingID, _ := strconv.Atoi(trainingIDStr)

	// Получаем userID через unified auth (initData или browser session)
	userID, err := h.getUserIDFromRequest(w, r)
	if err != nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	student, err := h.studentService.GetStudentByUserID(userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"registered": false,
			"error":      "student not found",
		})
		return
	}

	attendance, _ := h.attendanceService.GetStudentAttendanceForTraining(int(student.ID), trainingID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"registered": attendance != nil && attendance.Status == "registered",
		"attendance": attendance,
	})
}

// getUserIDFromRequest extracts userID from access JWT cookie, silent-refreshes via refresh cookie, or legacy user_id.
func (h *Handler) getUserIDFromRequest(w http.ResponseWriter, r *http.Request) (int64, error) {
	if h.sessionService != nil {
		access := cookieValue(r, accessCookieName)
		refresh := cookieValue(r, refreshCookieName)
		userID, newAccess, newRefresh, rotated, err := h.sessionService.ResolveUserID(access, refresh)
		if err == nil {
			if rotated && w != nil {
				h.writeAuthCookies(w, newAccess, newRefresh)
			}
			return userID, nil
		}
	}

	if h.allowLegacyUserID {
		userIDStr := r.URL.Query().Get("user_id")
		if userIDStr != "" {
			userID, err := strconv.ParseInt(userIDStr, 10, 64)
			if err == nil {
				return userID, nil
			}
		}
	}

	return 0, fmt.Errorf("user not authenticated")
}

// MyRegistrationsAPI возвращает список записей ученика на тренировки
func (h *Handler) MyRegistrationsAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем userID через unified auth (initData или browser session)
	userID, err := h.getUserIDFromRequest(w, r)
	if err != nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Проверяем роль пользователя - только студенты
	user, err := h.userService.GetByID(userID)
	if err != nil {
		http.Error(w, "User not found: "+err.Error(), http.StatusBadRequest)
		return
	}

	if user.Role != "student" {
		http.Error(w, "This endpoint is only for students", http.StatusForbidden)
		return
	}

	// Получаем студента
	student, err := h.studentService.GetStudentByUserID(userID)
	if err != nil {
		http.Error(w, "Student not found: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Записи: прошедший месяц и ближайшие 90 дней
	start := time.Now().AddDate(0, 0, -30)
	end := time.Now().AddDate(0, 0, 90)

	attendances, err := h.attendanceService.GetAttendanceByStudent(int(student.ID), start, end)
	if err != nil {
		http.Error(w, "Error getting registrations: "+err.Error(), http.StatusInternalServerError)
		return
	}

	upcomingTrainings := make([]map[string]interface{}, 0)
	pastTrainings := make([]map[string]interface{}, 0)

	now := time.Now()
	for _, attendance := range attendances {
		training, err := h.scheduleService.GetTrainingByID(attendance.TrainingID)
		if err != nil || training == nil {
			continue
		}

		groupName := training.GroupName
		if groupName == "" {
			groupName = "Неизвестная группа"
		}

		coachName := training.CoachName
		if coachName == "" {
			coachName = "Неизвестный тренер"
		}

		status := attendance.Status
		if status == "" {
			status = "registered"
		}

		trainingData := map[string]interface{}{
			"attendance_id": attendance.ID,
			"training_id":   training.ID,
			"date":          training.TrainingDate.Format("2006-01-02"),
			"time":          training.StartTime.Format("15:04"),
			"end_time":      training.EndTime.Format("15:04"),
			"group_name":    groupName,
			"coach_name":    coachName,
			"description":   training.Description,
			"status":        status,
			"attended":      attendance.Attended,
			"notes":         attendance.Notes,
		}

		// Предстоящие — пока не закончилась тренировка (дата+время), не только date > now
		if !trainingEndTime(training).Before(now) && status == "registered" && !attendance.Attended {
			upcomingTrainings = append(upcomingTrainings, trainingData)
		} else {
			pastTrainings = append(pastTrainings, trainingData)
		}
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"upcoming": upcomingTrainings,
		"past":     pastTrainings,
	})
}

// MySubscriptionAPI возвращает информацию об абонементе ученика
func (h *Handler) MySubscriptionAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем userID через unified auth (initData или browser session)
	userID, err := h.getUserIDFromRequest(w, r)
	if err != nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Проверяем роль пользователя - только студенты
	user, err := h.userService.GetByID(userID)
	if err != nil {
		http.Error(w, "User not found: "+err.Error(), http.StatusBadRequest)
		return
	}

	if user.Role != "student" {
		http.Error(w, "This endpoint is only for students", http.StatusForbidden)
		return
	}

	// Получаем студента
	student, err := h.studentService.GetStudentByUserID(userID)
	if err != nil {
		http.Error(w, "Student not found: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Получаем абонементы
	subscriptions, err := h.subscriptionService.GetSubscriptionsByStudentID(student.ID)
	if err != nil {
		http.Error(w, "Error getting subscriptions: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var activeSubscriptions []map[string]interface{}
	var expiredSubscriptions []map[string]interface{}
	now := time.Now()

	for _, subscription := range subscriptions {
		subData := map[string]interface{}{
			"id":                subscription.ID,
			"total_lessons":     subscription.TotalLessons,
			"remaining_lessons": subscription.RemainingLessons,
			"used_lessons":      subscription.TotalLessons - subscription.RemainingLessons,
			"start_date":        subscription.StartDate.Format("2006-01-02"),
			"end_date":          subscription.EndDate.Format("2006-01-02"),
			"created_at":        subscription.CreatedAt.Format("2006-01-02"),
		}

		if subscription.RemainingLessons >= -2 && subscription.EndDate.After(now) {
			activeSubscriptions = append(activeSubscriptions, subData)
		} else {
			status := "used"
			if subscription.EndDate.Before(now) || subscription.EndDate.Equal(now) {
				status = "expired"
			}
			subData["status"] = status
			expiredSubscriptions = append(expiredSubscriptions, subData)
		}
	}

	response := map[string]interface{}{
		"active":     activeSubscriptions,
		"expired":    expiredSubscriptions,
		"has_active": len(activeSubscriptions) > 0,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// MyProfileAPI возвращает профиль пользователя (ученика или тренера)
func (h *Handler) MyProfileAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем userID через unified auth (initData или browser session)
	userID, err := h.getUserIDFromRequest(w, r)
	if err != nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Получаем пользователя
	user, err := h.userService.GetByID(userID)
	if err != nil {
		http.Error(w, "User not found: "+err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"user": map[string]interface{}{
			"id":            user.ID,
			"first_name":    user.FirstName,
			"last_name":     user.LastName,
			"username":      user.Username,
			"role":          user.Role,
			"registered_at": user.RegisteredAt.Format("2006-01-02"),
		},
	}

	// В зависимости от роли получаем соответствующую информацию
	if user.Role == "student" {
		// Получаем студента
		student, err := h.studentService.GetStudentByUserID(userID)
		if err != nil {
			http.Error(w, "Student not found: "+err.Error(), http.StatusBadRequest)
			return
		}

		response["student"] = map[string]interface{}{
			"id":             student.ID,
			"athletic_title": student.AtleticTitle,
			"created_at":     student.CreatedAt.Format("2006-01-02"),
		}
	} else if user.Role == "coach" {
		// Получаем тренера
		coach, err := h.coachService.GetCoachByUserID(userID)
		if err != nil {
			http.Error(w, "Coach not found: "+err.Error(), http.StatusBadRequest)
			return
		}

		response["coach"] = map[string]interface{}{
			"id":          coach.ID,
			"specialty":   coach.Specialty,
			"experience":  coach.Experience,
			"description": coach.Description,
			"created_at":  coach.CreatedAt.Format("2006-01-02"),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateProfileAPI обновляет профиль пользователя
func (h *Handler) UpdateProfileAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем userID через unified auth (initData или browser session)
	userID, err := h.getUserIDFromRequest(w, r)
	if err != nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Получаем пользователя
	user, err := h.userService.GetByID(userID)
	if err != nil {
		http.Error(w, "User not found: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Парсим JSON тело запроса
	var requestData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Подготавливаем данные для обновления пользователя
	var firstName, lastName *string
	if fn, ok := requestData["first_name"].(string); ok && fn != "" {
		firstName = &fn
	}
	if ln, ok := requestData["last_name"].(string); ok && ln != "" {
		lastName = &ln
	}

	// Обновляем данные пользователя (имя и фамилия)
	if firstName != nil || lastName != nil {
		if err := h.userService.UpdateProfile(userID, firstName, lastName); err != nil {
			http.Error(w, "Error updating user: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Обновляем данные в зависимости от роли
	if user.Role == "student" {
		// Получаем студента
		student, err := h.studentService.GetStudentByUserID(userID)
		if err != nil {
			http.Error(w, "Student not found: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Обновляем athletic_title если передан
		if at, ok := requestData["athletic_title"].(string); ok {
			var athleticTitle *string
			if at != "" {
				athleticTitle = &at
			}
			if err := h.studentService.UpdateAthleticTitle(student.ID, athleticTitle); err != nil {
				http.Error(w, "Error updating student: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}
	} else if user.Role == "coach" {
		// Получаем тренера
		coach, err := h.coachService.GetCoachByUserID(userID)
		if err != nil {
			http.Error(w, "Coach not found: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Подготавливаем данные для обновления тренера
		var specialty, experience, description *string
		if sp, ok := requestData["specialty"].(string); ok {
			if sp != "" {
				specialty = &sp
			}
		}
		if exp, ok := requestData["experience"].(string); ok {
			if exp != "" {
				experience = &exp
			}
		}
		if desc, ok := requestData["description"].(string); ok {
			if desc != "" {
				description = &desc
			}
		}

		// Обновляем профиль тренера если переданы данные
		if specialty != nil || experience != nil || description != nil {
			if err := h.coachService.UpdateCoachProfile(coach.ID, specialty, experience, description); err != nil {
				http.Error(w, "Error updating coach: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	// Получаем обновленного пользователя
	updatedUser, err := h.userService.GetByID(userID)
	if err != nil {
		http.Error(w, "Error getting updated user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Формируем ответ с обновленным профилем
	response := map[string]interface{}{
		"user": map[string]interface{}{
			"id":            updatedUser.ID,
			"first_name":    updatedUser.FirstName,
			"last_name":     updatedUser.LastName,
			"username":      updatedUser.Username,
			"role":          updatedUser.Role,
			"registered_at": updatedUser.RegisteredAt.Format("2006-01-02"),
		},
	}

	// В зависимости от роли добавляем соответствующую информацию
	if updatedUser.Role == "student" {
		student, err := h.studentService.GetStudentByUserID(userID)
		if err == nil && student != nil {
			response["student"] = map[string]interface{}{
				"id":             student.ID,
				"athletic_title": student.AtleticTitle,
				"created_at":     student.CreatedAt.Format("2006-01-02"),
			}
		}
	} else if updatedUser.Role == "coach" {
		coach, err := h.coachService.GetCoachByUserID(userID)
		if err == nil && coach != nil {
			response["coach"] = map[string]interface{}{
				"id":          coach.ID,
				"specialty":   coach.Specialty,
				"experience":  coach.Experience,
				"description": coach.Description,
				"created_at":  coach.CreatedAt.Format("2006-01-02"),
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AllStudentsSubscriptionsAPI возвращает все абонементы всех учеников (только для тренеров)
func (h *Handler) AllStudentsSubscriptionsAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем userID через unified auth (initData или browser session)
	userID, err := h.getUserIDFromRequest(w, r)
	if err != nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Проверяем, что пользователь - тренер
	user, err := h.userService.GetByID(userID)
	if err != nil {
		http.Error(w, "User not found: "+err.Error(), http.StatusBadRequest)
		return
	}

	if user.Role != "coach" {
		http.Error(w, "This endpoint is only for coaches", http.StatusForbidden)
		return
	}

	// Получаем всех студентов
	students, err := h.userService.GetAllStudents()
	if err != nil {
		http.Error(w, "Error getting students: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var studentsWithSubscriptions []map[string]interface{}
	now := time.Now()

	for _, studentUser := range students {
		// Получаем студента по userID
		student, err := h.studentService.GetStudentByUserID(studentUser.ID)
		if err != nil {
			// Пропускаем, если студент не найден
			continue
		}

		// Получаем абонементы студента
		subscriptions, err := h.subscriptionService.GetSubscriptionsByStudentID(student.ID)
		if err != nil {
			// Пропускаем, если не удалось получить абонементы
			continue
		}

		// Формируем данные об абонементах
		var subscriptionData []map[string]interface{}
		for _, subscription := range subscriptions {
			subData := map[string]interface{}{
				"id":                subscription.ID,
				"total_lessons":     subscription.TotalLessons,
				"remaining_lessons": subscription.RemainingLessons,
				"used_lessons":      subscription.TotalLessons - subscription.RemainingLessons,
				"start_date":        subscription.StartDate.Format("2006-01-02"),
				"end_date":          subscription.EndDate.Format("2006-01-02"),
				"created_at":        subscription.CreatedAt.Format("2006-01-02"),
			}

			// Определяем статус
			// Активен при remaining_lessons >= -2 и end_date > now
			var status string
			if subscription.EndDate.Before(now) || subscription.EndDate.Equal(now) {
				status = "expired"
			} else if subscription.RemainingLessons >= -2 {
				status = "active"
			} else {
				status = "used"
			}
			subData["status"] = status

			subscriptionData = append(subscriptionData, subData)
		}

		// Подсчитываем общее количество занятий из активных абонементов
		totalLessonsCount := 0
		hasActiveSubscription := false
		for _, subscription := range subscriptions {
			if subscription.RemainingLessons >= -2 && subscription.EndDate.After(now) {
				hasActiveSubscription = true
				// Для счетчика не показываем отрицательные "доступные занятия".
				if subscription.RemainingLessons > 0 {
					totalLessonsCount += subscription.RemainingLessons
				}
			}
		}

		// Добавляем студента в список только если у него есть активные абонементы
		if hasActiveSubscription {
			studentData := map[string]interface{}{
				"user_id":             studentUser.ID,
				"student_id":          student.ID,
				"name":                formatStudentDisplayName(studentUser),
				"subscriptions":       subscriptionData,
				"total_lessons_count": totalLessonsCount,
			}

			studentsWithSubscriptions = append(studentsWithSubscriptions, studentData)
		}
	}

	response := map[string]interface{}{
		"students": studentsWithSubscriptions,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// sendJSONError отправляет JSON ошибку
func (h *Handler) sendJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": message,
	})
}

// AddLessonsAPI добавляет занятия к абонементу (только для тренеров)
func (h *Handler) AddLessonsAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Получаем userID и проверяем, что это тренер
	userID, err := h.getUserIDFromRequest(w, r)
	if err != nil {
		h.sendJSONError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	user, err := h.userService.GetByID(userID)
	if err != nil {
		h.sendJSONError(w, http.StatusBadRequest, "User not found: "+err.Error())
		return
	}

	if user.Role != "coach" {
		h.sendJSONError(w, http.StatusForbidden, "Access denied: only coaches can modify subscriptions")
		return
	}

	// Парсим JSON тело запроса
	var requestData struct {
		SubscriptionID int64 `json:"subscription_id"`
		Count          int   `json:"count"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		h.sendJSONError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	if requestData.Count <= 0 {
		h.sendJSONError(w, http.StatusBadRequest, "Count must be greater than 0")
		return
	}

	// Добавляем занятия
	if err := h.subscriptionService.AddLessons(requestData.SubscriptionID, requestData.Count); err != nil {
		h.sendJSONError(w, http.StatusInternalServerError, "Error adding lessons: "+err.Error())
		return
	}

	// Получаем обновленный абонемент
	allSubs, err := h.subscriptionService.GetAll()
	if err == nil {
		now := time.Now()
		for _, sub := range allSubs {
			if sub.ID == requestData.SubscriptionID {
				// Определяем статус
				var status string
				if sub.EndDate.Before(now) || sub.EndDate.Equal(now) {
					status = "expired"
				} else if sub.RemainingLessons >= -2 {
					status = "active"
				} else {
					status = "used"
				}

				response := map[string]interface{}{
					"id":                sub.ID,
					"total_lessons":     sub.TotalLessons,
					"remaining_lessons": sub.RemainingLessons,
					"used_lessons":      sub.TotalLessons - sub.RemainingLessons,
					"start_date":        sub.StartDate.Format("2006-01-02"),
					"end_date":          sub.EndDate.Format("2006-01-02"),
					"created_at":        sub.CreatedAt.Format("2006-01-02"),
					"status":            status,
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
				return
			}
		}
	}

	// Если не нашли, возвращаем успешный ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Lessons added successfully",
	})
}

// RemoveLessonsAPI снимает занятия с абонемента (только для тренеров)
func (h *Handler) RemoveLessonsAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Получаем userID и проверяем, что это тренер
	userID, err := h.getUserIDFromRequest(w, r)
	if err != nil {
		h.sendJSONError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	user, err := h.userService.GetByID(userID)
	if err != nil {
		h.sendJSONError(w, http.StatusBadRequest, "User not found: "+err.Error())
		return
	}

	if user.Role != "coach" {
		h.sendJSONError(w, http.StatusForbidden, "Access denied: only coaches can modify subscriptions")
		return
	}

	// Парсим JSON тело запроса
	var requestData struct {
		SubscriptionID int64 `json:"subscription_id"`
		Count          int   `json:"count"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		h.sendJSONError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	if requestData.Count <= 0 {
		h.sendJSONError(w, http.StatusBadRequest, "Count must be greater than 0")
		return
	}

	// Снимаем занятия
	if err := h.subscriptionService.RemoveLessons(requestData.SubscriptionID, requestData.Count); err != nil {
		h.sendJSONError(w, http.StatusInternalServerError, "Error removing lessons: "+err.Error())
		return
	}

	// Получаем обновленный абонемент
	allSubs, err := h.subscriptionService.GetAll()
	if err == nil {
		now := time.Now()
		for _, sub := range allSubs {
			if sub.ID == requestData.SubscriptionID {
				// Определяем статус
				var status string
				if sub.EndDate.Before(now) || sub.EndDate.Equal(now) {
					status = "expired"
				} else if sub.RemainingLessons >= -2 {
					status = "active"
				} else {
					status = "used"
				}

				response := map[string]interface{}{
					"id":                sub.ID,
					"total_lessons":     sub.TotalLessons,
					"remaining_lessons": sub.RemainingLessons,
					"used_lessons":      sub.TotalLessons - sub.RemainingLessons,
					"start_date":        sub.StartDate.Format("2006-01-02"),
					"end_date":          sub.EndDate.Format("2006-01-02"),
					"created_at":        sub.CreatedAt.Format("2006-01-02"),
					"status":            status,
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
				return
			}
		}
	}

	// Если не нашли, возвращаем успешный ответ
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Lessons removed successfully",
	})
}

// SetSubscriptionEndDateAPI меняет дату окончания абонемента (только для тренеров).
func (h *Handler) SetSubscriptionEndDateAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, err := h.getUserIDFromRequest(w, r)
	if err != nil {
		h.sendJSONError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	user, err := h.userService.GetByID(userID)
	if err != nil {
		h.sendJSONError(w, http.StatusBadRequest, "User not found: "+err.Error())
		return
	}
	if user.Role != "coach" {
		h.sendJSONError(w, http.StatusForbidden, "Access denied: only coaches can modify subscriptions")
		return
	}

	var requestData struct {
		SubscriptionID int64  `json:"subscription_id"`
		EndDate        string `json:"end_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		h.sendJSONError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}
	endDate, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(requestData.EndDate), time.Local)
	if err != nil {
		h.sendJSONError(w, http.StatusBadRequest, "invalid end_date, use YYYY-MM-DD")
		return
	}
	if err := h.subscriptionService.SetEndDate(requestData.SubscriptionID, endDate); err != nil {
		h.sendJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	allSubs, err := h.subscriptionService.GetAll()
	if err == nil {
		now := time.Now()
		for _, sub := range allSubs {
			if sub.ID != requestData.SubscriptionID {
				continue
			}
			status := "active"
			if sub.EndDate.Before(now) || sub.EndDate.Equal(now) {
				status = "expired"
			} else if sub.RemainingLessons < -2 {
				status = "used"
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":                sub.ID,
				"total_lessons":     sub.TotalLessons,
				"remaining_lessons": sub.RemainingLessons,
				"used_lessons":      sub.TotalLessons - sub.RemainingLessons,
				"start_date":        sub.StartDate.Format("2006-01-02"),
				"end_date":          sub.EndDate.Format("2006-01-02"),
				"created_at":        sub.CreatedAt.Format("2006-01-02"),
				"status":            status,
			})
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "End date updated",
	})
}
