package web

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"spectrum-club-bot/internal/models"
	"spectrum-club-bot/internal/service"
)

type Handler struct {
	scheduleService   service.TrainingScheduleService
	coachService      service.CoachService
	attendanceService service.AttendanceService
	studentService    service.StudentService
	userService       service.UserService
	botToken          string // Для проверки Telegram WebApp initData
}

func NewHandler(
	scheduleService service.TrainingScheduleService,
	coachService service.CoachService,
	attendanceService service.AttendanceService,
	studentService service.StudentService,
	userService service.UserService,
	botToken string,
) *Handler {
	return &Handler{
		scheduleService:   scheduleService,
		coachService:      coachService,
		attendanceService: attendanceService,
		studentService:    studentService,
		userService:       userService,
		botToken:          botToken,
	}
}

// Calendar method removed - теперь используется Angular фронтенд
// CalendarAPI используется вместо Calendar для возврата JSON данных

// CalendarAPI возвращает данные календаря в формате JSON
func (h *Handler) CalendarAPI(w http.ResponseWriter, r *http.Request) {
	// Получаем userID из initData или query параметра (fallback)
	userID, userErr := h.getUserIDFromRequest(r)
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

		// Добавляем тренировки для этого дня
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
					if training.MaxParticipants != nil && *training.MaxParticipants > 0 {
						maxParticipants := *training.MaxParticipants
						if len(participants) < maxParticipants {
							canRegister = true
						}
					} else {
						canRegister = true
					}
				}
			}
		}

		maxParticipants := 0
		if training.MaxParticipants != nil {
			maxParticipants = *training.MaxParticipants
		}

		grouped[dateKey] = append(grouped[dateKey], TrainingViewJSON{
			ID:               training.ID,
			GroupName:        training.GroupName,
			StartTime:        training.StartTime.Format("15:04"),
			EndTime:          training.EndTime.Format("15:04"),
			CoachName:        training.CoachName,
			Participants:     len(participants),
			ParticipantNames: participantNames,
			MaxParticipants:  maxParticipants,
			CanRegister:      canRegister,
			IsRegistered:     isRegistered,
			IsFull:           training.MaxParticipants != nil && len(participants) >= *training.MaxParticipants,
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

// Структуры данных для шаблонов
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

func (h *Handler) getWeekDayHeaders() []WeekDayHeader {
	return []WeekDayHeader{
		{Name: "ПН", Day: ""},
		{Name: "ВТ", Day: ""},
		{Name: "СР", Day: ""},
		{Name: "ЧТ", Day: ""},
		{Name: "ПТ", Day: ""},
		{Name: "СБ", Day: ""},
		{Name: "ВС", Day: ""},
	}
}

func (h *Handler) generateTimeSlots() []time.Time {
	var slots []time.Time
	for hour := 8; hour <= 22; hour++ {
		slots = append(slots, time.Date(0, 1, 1, hour, 0, 0, 0, time.UTC))
	}
	return slots
}

func (h *Handler) prepareMonthView(date time.Time, trainings []models.TrainingSchedule, userID string) []CalendarDay {
	firstDay := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.Local)

	// Начинаем с понедельника недели, содержащей первый день
	weekday := int(firstDay.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	startDay := firstDay.AddDate(0, 0, -(weekday - 1))

	var days []CalendarDay

	// 6 недель = 42 дня
	for i := 0; i < 42; i++ {
		day := startDay.AddDate(0, 0, i)

		dayData := CalendarDay{
			Date:         day,
			IsToday:      day.Format("2006-01-02") == time.Now().Format("2006-01-02"),
			IsOtherMonth: day.Month() != date.Month(),
		}

		// Добавляем тренировки для этого дня
		var events []CalendarEvent
		for _, training := range trainings {
			trainingDate := training.TrainingDate.Format("2006-01-02")
			dayStr := day.Format("2006-01-02")

			if trainingDate == dayStr {
				// Получаем участников
				// participants, _ := h.attendanceService.GetParticipants(training.ID)
				// fmt.Println(participants)
				events = append(events, CalendarEvent{
					ID:    training.ID,
					Title: training.GroupName,
					Time: fmt.Sprintf("%s - %s",
						training.StartTime.Format("15:04"),
						training.EndTime.Format("15:04")),
					Coach:      training.CoachName,
					Top:        30 + float64(len(events))*30,
					Height:     25,
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

func (h *Handler) prepareWeekView(startDate time.Time, trainings []models.TrainingSchedule, userID string) []WeekDayData {
	var weekDays []WeekDayData

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

		dayData := WeekDayData{
			Name:    dayName,
			Day:     day.Day(),
			Date:    day,
			IsToday: day.Format("2006-01-02") == time.Now().Format("2006-01-02"),
		}

		// Добавляем тренировки для этого дня
		var events []CalendarEvent
		for _, training := range trainings {
			trainingDate := training.TrainingDate.Format("2006-01-02")
			dayStr := day.Format("2006-01-02")

			if trainingDate == dayStr {
				// Рассчитываем позицию и высоту
				startMinutes := training.StartTime.Hour()*60 + training.StartTime.Minute()
				endMinutes := training.EndTime.Hour()*60 + training.EndTime.Minute()

				// Позиция относительно 8:00 утра (480 минут)
				top := float64(startMinutes-480) * 1.0 // 1px на минуту
				height := float64(endMinutes-startMinutes) * 1.0

				events = append(events, CalendarEvent{
					ID:    training.ID,
					Title: training.GroupName,
					Time: fmt.Sprintf("%s - %s",
						training.StartTime.Format("15:04"),
						training.EndTime.Format("15:04")),
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

func (h *Handler) prepareDayView(date time.Time, trainings []models.TrainingSchedule, userID string) []CalendarEvent {
	var events []CalendarEvent

	dateStr := date.Format("2006-01-02")

	for _, training := range trainings {
		trainingDate := training.TrainingDate.Format("2006-01-02")

		if trainingDate == dateStr {
			// Рассчитываем позицию и высоту
			startMinutes := training.StartTime.Hour()*60 + training.StartTime.Minute()
			endMinutes := training.EndTime.Hour()*60 + training.EndTime.Minute()

			top := float64(startMinutes-480) * 1.0 // 1px на минуту
			height := float64(endMinutes-startMinutes) * 1.0

			events = append(events, CalendarEvent{
				ID:    training.ID,
				Title: training.GroupName,
				Time: fmt.Sprintf("%s - %s",
					training.StartTime.Format("15:04"),
					training.EndTime.Format("15:04")),
				Coach:      training.CoachName,
				Top:        top,
				Height:     height,
				ColorIndex: h.getColorIndex(training.GroupID),
				UserID:     userID,
			})
		}
	}

	return events
}

func (h *Handler) prepareScheduleView(trainings []models.TrainingSchedule, userID string, isCoach bool) []ScheduleDay {
	// Группируем по дням
	grouped := make(map[string][]TrainingView)

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
		// Получаем участников

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
					if training.MaxParticipants != nil && *training.MaxParticipants > 0 {
						maxParticipants := *training.MaxParticipants
						if len(participants) < maxParticipants {
							canRegister = true
						}
					} else {
						canRegister = true
					}
				}
			}
		}

		grouped[dateKey] = append(grouped[dateKey], TrainingView{
			ID:               training.ID,
			GroupName:        training.GroupName,
			StartTime:        training.StartTime.Format("15:04"),
			EndTime:          training.EndTime.Format("15:04"),
			CoachName:        training.CoachName,
			Participants:     len(participants),
			ParticipantNames: participantNames, // Добавляем имена
			MaxParticipants: func() int {
				if training.MaxParticipants != nil {
					return *training.MaxParticipants
				}
				return 0
			}(),
			CanRegister:  canRegister,
			IsRegistered: isRegistered,
			IsFull:       training.MaxParticipants != nil && len(participants) >= *training.MaxParticipants,
			ColorIndex:   h.getColorIndex(training.GroupID),
		})
	}

	// Преобразуем в массив и сортируем по дате
	var scheduleDays []ScheduleDay
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

		scheduleDays = append(scheduleDays, ScheduleDay{
			Date:      formattedDate,
			Trainings: trainings,
		})
	}

	return scheduleDays
}

func (h *Handler) getColorIndex(groupID int) int {
	if groupID <= 0 {
		return 1
	}
	return (groupID % 5) + 1
}

// Структуры для спискового вида
type TrainingView struct {
	ID               int
	GroupName        string
	StartTime        string
	EndTime          string
	CoachName        string
	Participants     int
	ParticipantNames string // Добавьте это поле
	MaxParticipants  int
	CanRegister      bool
	IsRegistered     bool
	IsFull           bool
	ColorIndex       int
}

type ScheduleDay struct {
	Date      string
	Trainings []TrainingView
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
	userID, err := h.getUserIDFromRequest(r)
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
	// 3. Есть свободные места (если установлен лимит)
	// 4. Пользователь не записан (если userID передан, иначе считаем что не записан)
	if !isCoach && !isRegistered && trainingDateTime.After(now) {
		if training.MaxParticipants != nil && *training.MaxParticipants > 0 {
			maxParticipants := *training.MaxParticipants
			if len(participants) < maxParticipants {
				canRegister = true
			} else {
				isFull = true
			}
		} else {
			// Если лимита нет, всегда можно записаться (если тренировка в будущем и пользователь - студент)
			canRegister = true
		}
	}

	// Формируем ответ
	response := map[string]interface{}{
		"training": map[string]interface{}{
			"id":               training.ID,
			"group_name":       training.GroupName,
			"training_date":    training.TrainingDate.Format("2006-01-02"),
			"start_time":       training.StartTime.Format("15:04"),
			"end_time":         training.EndTime.Format("15:04"),
			"coach_name":       training.CoachName,
			"description":      training.Description,
			"max_participants": training.MaxParticipants,
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

	// Получаем userID из initData (безопасный способ) или из формы (fallback)
	userID, authErr := h.getUserIDFromRequest(r)
	if authErr != nil {
		log.Printf("[RegisterForTraining] ОШИБКА аутентификации через initData: %v", authErr)
		log.Printf("[RegisterForTraining] initData в заголовке: %v (длина: %d)",
			r.Header.Get("X-Telegram-Init-Data") != "", len(r.Header.Get("X-Telegram-Init-Data")))

		// Fallback: из формы (для обратной совместимости)
		userIDStr := r.FormValue("user_id")
		log.Printf("[RegisterForTraining] Пробуем fallback: user_id из формы: %s", userIDStr)

		if userIDStr == "" {
			log.Printf("[RegisterForTraining] ❌ Аутентификация не удалась: нет ни initData, ни user_id")
			http.Error(w, "Authentication required: Необходимо войти в систему. Пожалуйста, откройте календарь через Telegram бота.", http.StatusUnauthorized)
			return
		}
		userID, err = strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}
		log.Printf("[RegisterForTraining] ✅ Использован fallback: userID из формы: %d", userID)
	} else {
		log.Printf("[RegisterForTraining] ✅ Аутентификация успешна через initData, userID: %d", userID)
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

	// Проверяем, существует ли тренировка
	training, err := h.scheduleService.GetTrainingByID(trainingID)
	if err != nil {
		http.Error(w, "Training not found: "+err.Error(), http.StatusNotFound)
		return
	}

	// Проверяем, не прошла ли тренировка
	if training.TrainingDate.Before(time.Now()) {
		http.Error(w, "Training has already passed", http.StatusBadRequest)
		return
	}

	// Проверяем, есть ли места
	participants, err := h.attendanceService.GetParticipants(trainingID)
	if err == nil && training.MaxParticipants != nil {
		if len(participants) >= *training.MaxParticipants {
			http.Error(w, "No available spots", http.StatusBadRequest)
			return
		}
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

	// Получаем userID из initData (безопасный способ) или из формы (fallback)
	userID, authErr := h.getUserIDFromRequest(r)
	if authErr != nil {
		// Fallback: из формы (для обратной совместимости)
		userIDStr := r.FormValue("user_id")
		if userIDStr == "" {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}
		userID, err = strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}
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

	// Проверяем, можно ли отменить (например, тренировка еще не прошла)
	training, err := h.scheduleService.GetTrainingByID(trainingID)
	if err == nil {
		// Можно отменить только если тренировка в будущем
		if training.TrainingDate.Before(time.Now()) {
			http.Error(w, "Cannot cancel past training", http.StatusBadRequest)
			return
		}
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
	userID, authErr := h.getUserIDFromRequest(r)
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

		err := h.attendanceService.MarkAttendance(
			requestData.TrainingID,
			studentID,
			int(coach.ID),
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

	// Получаем userID из initData (безопасный способ) или из query параметра (fallback)
	userID, err := h.getUserIDFromRequest(r)
	if err != nil {
		// Fallback: из query параметра (для обратной совместимости)
		userIDStr := r.URL.Query().Get("user_id")
		if userIDStr == "" {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}
		userID, _ = strconv.ParseInt(userIDStr, 10, 64)
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

// verifyTelegramWebAppData проверяет подлинность Telegram WebApp initData
// Возвращает telegram_id если данные валидны, иначе ошибку
func (h *Handler) verifyTelegramWebAppData(initData string) (int64, error) {
	log.Printf("[verifyTelegramWebAppData] Начало проверки initData, длина: %d", len(initData))

	if h.botToken == "" {
		log.Printf("[verifyTelegramWebAppData] ОШИБКА: bot token не настроен")
		return 0, fmt.Errorf("bot token not configured")
	}

	// Парсим initData
	parsed, err := url.ParseQuery(initData)
	if err != nil {
		log.Printf("[verifyTelegramWebAppData] ОШИБКА: неверный формат initData: %v", err)
		return 0, fmt.Errorf("invalid initData format: %v", err)
	}

	log.Printf("[verifyTelegramWebAppData] initData распарсен, параметров: %d", len(parsed))

	// Извлекаем hash и остальные параметры
	hash := parsed.Get("hash")
	if hash == "" {
		log.Printf("[verifyTelegramWebAppData] ОШИБКА: hash не найден в initData")
		return 0, fmt.Errorf("hash not found in initData")
	}

	log.Printf("[verifyTelegramWebAppData] hash найден: %s", hash[:min(20, len(hash))]+"...")

	// Удаляем hash из параметров для проверки
	parsed.Del("hash")

	// Сортируем параметры по ключу
	keys := make([]string, 0, len(parsed))
	for k := range parsed {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Формируем data-check-string
	var dataCheckString strings.Builder
	for i, k := range keys {
		if i > 0 {
			dataCheckString.WriteString("\n")
		}
		dataCheckString.WriteString(k)
		dataCheckString.WriteString("=")
		values := parsed[k]
		if len(values) > 0 {
			dataCheckString.WriteString(values[0])
		}
	}

	// Вычисляем секретный ключ
	secretKey := hmac.New(sha256.New, []byte("WebAppData"))
	secretKey.Write([]byte(h.botToken))
	secretKeyBytes := secretKey.Sum(nil)

	// Вычисляем HMAC
	mac := hmac.New(sha256.New, secretKeyBytes)
	mac.Write([]byte(dataCheckString.String()))
	expectedHash := hex.EncodeToString(mac.Sum(nil))

	// Сравниваем хеши
	if hash != expectedHash {
		return 0, fmt.Errorf("invalid hash: data not from Telegram")
	}

	// Извлекаем user из user параметра
	userStr := parsed.Get("user")
	if userStr == "" {
		return 0, fmt.Errorf("user not found in initData")
	}

	// Парсим JSON user
	var userData struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal([]byte(userStr), &userData); err != nil {
		return 0, fmt.Errorf("invalid user data: %v", err)
	}

	return userData.ID, nil
}

// min возвращает минимальное из двух чисел
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// AuthAPI проверяет Telegram WebApp initData и возвращает userID из базы
func (h *Handler) AuthAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем initData из заголовка или тела запроса
	initData := r.Header.Get("X-Telegram-Init-Data")
	if initData == "" {
		// Пробуем получить из тела запроса
		if err := r.ParseForm(); err == nil {
			initData = r.FormValue("initData")
		}
	}

	if initData == "" {
		http.Error(w, "Missing initData", http.StatusBadRequest)
		return
	}

	// Проверяем initData и получаем telegram_id
	telegramID, err := h.verifyTelegramWebAppData(initData)
	if err != nil {
		http.Error(w, "Invalid initData: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Находим user в базе по telegram_id
	user, err := h.userService.GetByTelegramID(telegramID)
	if err != nil {
		http.Error(w, "User not found: "+err.Error(), http.StatusNotFound)
		return
	}

	// Возвращаем userID
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id":     user.ID,
		"telegram_id": telegramID,
	})
}

// getUserIDFromRequest извлекает userID из запроса через initData или query параметр (fallback)
func (h *Handler) getUserIDFromRequest(r *http.Request) (int64, error) {
	// Пробуем получить из initData (безопасный способ)
	initData := r.Header.Get("X-Telegram-Init-Data")
	if initData != "" {
		log.Printf("[getUserIDFromRequest] initData получен, длина: %d", len(initData))
		log.Printf("[getUserIDFromRequest] initData (первые 100 символов): %s",
			func() string {
				if len(initData) > 100 {
					return initData[:100] + "..."
				}
				return initData
			}())

		telegramID, err := h.verifyTelegramWebAppData(initData)
		if err == nil {
			log.Printf("[getUserIDFromRequest] initData проверен успешно, telegramID: %d", telegramID)
			user, err := h.userService.GetByTelegramID(telegramID)
			if err == nil {
				log.Printf("[getUserIDFromRequest] Пользователь найден, userID: %d", user.ID)
				return user.ID, nil
			} else {
				log.Printf("[getUserIDFromRequest] Пользователь не найден по telegramID %d: %v", telegramID, err)
			}
		} else {
			log.Printf("[getUserIDFromRequest] Ошибка проверки initData: %v", err)
		}
	} else {
		log.Printf("[getUserIDFromRequest] initData не найден в заголовках")
	}

	// Fallback: из query параметра (для обратной совместимости, но небезопасно)
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr != "" {
		log.Printf("[getUserIDFromRequest] Используется fallback: user_id из query параметра: %s", userIDStr)
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err == nil {
			return userID, nil
		}
	}

	log.Printf("[getUserIDFromRequest] Аутентификация не удалась: initData пустой и user_id не найден")
	return 0, fmt.Errorf("user not authenticated")
}
