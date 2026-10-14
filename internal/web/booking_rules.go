package web

import (
	"fmt"
	"time"

	"spectrum-club-bot/internal/models"
)

const cancelRegisterHoursBefore = 2

func trainingStartTime(training *models.TrainingSchedule) time.Time {
	return time.Date(
		training.TrainingDate.Year(),
		training.TrainingDate.Month(),
		training.TrainingDate.Day(),
		training.StartTime.Hour(),
		training.StartTime.Minute(),
		training.StartTime.Second(),
		0,
		training.TrainingDate.Location(),
	)
}

func trainingEndTime(training *models.TrainingSchedule) time.Time {
	return time.Date(
		training.TrainingDate.Year(),
		training.TrainingDate.Month(),
		training.TrainingDate.Day(),
		training.EndTime.Hour(),
		training.EndTime.Minute(),
		training.EndTime.Second(),
		0,
		training.TrainingDate.Location(),
	)
}

// errIfTooLateForBooking returns error if now is within 2h of start (or after).
func errIfTooLateForBooking(training *models.TrainingSchedule, action string) error {
	deadline := trainingStartTime(training).Add(-cancelRegisterHoursBefore * time.Hour)
	if !time.Now().Before(deadline) {
		return fmt.Errorf("%s недоступна менее чем за %d часа до начала тренировки", action, cancelRegisterHoursBefore)
	}
	return nil
}
