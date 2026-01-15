package web

import (
	"encoding/json"
	"fmt"
	"net/http"
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
}

func NewHandler(
	scheduleService service.TrainingScheduleService,
	coachService service.CoachService,
	attendanceService service.AttendanceService,
	studentService service.StudentService,
	userService service.UserService,
) *Handler {
	return &Handler{
		scheduleService:   scheduleService,
		coachService:      coachService,
		attendanceService: attendanceService,
		studentService:    studentService,
		userService:       userService,
	}
}

// Calendar method removed - теперь используется Angular фронтенд
// CalendarAPI используется вместо Calendar для возврата JSON данных

// CalendarAPI возвращает данные календаря в формате JSON
func (h *Handler) CalendarAPI(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
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

	// Получаем пользователя
	var isCoach bool
	var userName string
	var userID int64

	if userIDStr != "" {
		var err error
		userID, err = strconv.ParseInt(userIDStr, 10, 64)
		if err == nil {
			user, err := h.userService.GetByID(userID)
			if err == nil {
				isCoach = user.Role == "coach"
				userName = user.FirstName + " " + user.LastName
			}
		}
	}

	// Получаем тренировки
	var trainings []models.TrainingSchedule
	var err error

	if isCoach && userIDStr != "" {
		// Для тренера получаем его тренировки
		coach, coachErr := h.coachService.GetCoachByUserID(userID)
		if coachErr == nil {
			trainings, err = h.scheduleService.GetCoachSchedule(coach.ID, startDate, endDate)
		}
	} else {
		// Для студента или незарегистрированного пользователя получаем все тренировки
		trainings, err = h.scheduleService.GetTrainingsByDateRange(startDate, endDate)
	}

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
				// Приводим к локальному времени для правильного извлечения часов и минут
				startTime := training.StartTime.In(time.Local)
				endTime := training.EndTime.In(time.Local)

				// Извлекаем часы и минуты из локального времени
				startHour := startTime.Hour()
				startMinute := startTime.Minute()
				endHour := endTime.Hour()
				endMinute := endTime.Minute()

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
			// Приводим к локальному времени для правильного извлечения часов и минут
			startTime := training.StartTime.In(time.Local)
			endTime := training.EndTime.In(time.Local)

			// Извлекаем часы и минуты из локального времени
			startHour := startTime.Hour()
			startMinute := startTime.Minute()
			endHour := endTime.Hour()
			endMinute := endTime.Minute()

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

				// Проверяем, можно ли записаться
				if !isRegistered && training.TrainingDate.After(time.Now()) {
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

				// Проверяем, можно ли записаться
				if !isRegistered && training.TrainingDate.After(time.Now()) {
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

	// Получаем user_id
	userIDStr := r.URL.Query().Get("user_id")
	var userID int64
	var isCoach bool
	var userName string

	if userIDStr != "" {
		var err error
		userID, err = strconv.ParseInt(userIDStr, 10, 64)
		if err == nil {
			user, err := h.userService.GetByID(userID)
			if err == nil {
				isCoach = user.Role == "coach"
				userName = user.FirstName + " " + user.LastName
			}
		}
	}
	fmt.Println(userName)
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
	isPast := time.Now().After(training.TrainingDate)

	if userIDStr != "" && !isCoach {
		student, err := h.studentService.GetStudentByUserID(userID)
		if err == nil {
			att, _ := h.attendanceService.GetStudentAttendanceForTraining(int(student.ID), trainingID)
			isRegistered = att != nil && att.Status == "registered"

			// Проверяем, можно ли записаться
			if !isRegistered && training.TrainingDate.After(time.Now()) {
				if training.MaxParticipants != nil && *training.MaxParticipants > 0 {
					maxParticipants := *training.MaxParticipants
					if len(participants) < maxParticipants {
						canRegister = true
					} else {
						isFull = true
					}
				} else {
					canRegister = true
				}
			}
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
		"participants":       participants,
		"participants_count": len(participants),
		"is_coach":           isCoach,
		"is_registered":      isRegistered,
		"can_register":       canRegister,
		"is_full":            isFull,
		"is_past":            isPast,
		"current_time":       time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RegisterForTraining обрабатывает запись на тренировку
func (h *Handler) RegisterForTraining(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем параметры из формы
	trainingIDStr := r.FormValue("training_id")
	userIDStr := r.FormValue("user_id")

	if trainingIDStr == "" || userIDStr == "" {
		http.Error(w, "Missing parameters", http.StatusBadRequest)
		return
	}

	trainingID, err := strconv.Atoi(trainingIDStr)
	if err != nil {
		http.Error(w, "Invalid training ID", http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
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

	// Получаем параметры из формы
	trainingIDStr := r.FormValue("training_id")
	userIDStr := r.FormValue("user_id")

	if trainingIDStr == "" || userIDStr == "" {
		http.Error(w, "Missing parameters", http.StatusBadRequest)
		return
	}

	trainingID, err := strconv.Atoi(trainingIDStr)
	if err != nil {
		http.Error(w, "Invalid training ID", http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
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

// CheckRegistration проверяет статус записи студента на тренировку
func (h *Handler) CheckRegistration(w http.ResponseWriter, r *http.Request) {
	trainingIDStr := r.URL.Query().Get("training_id")
	userIDStr := r.URL.Query().Get("user_id")

	if trainingIDStr == "" || userIDStr == "" {
		http.Error(w, "Missing parameters", http.StatusBadRequest)
		return
	}

	trainingID, _ := strconv.Atoi(trainingIDStr)
	userID, _ := strconv.ParseInt(userIDStr, 10, 64)

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
