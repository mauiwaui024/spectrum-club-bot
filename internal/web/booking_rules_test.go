package web

import (
	"testing"
	"time"

	"spectrum-club-bot/internal/models"
)

func TestErrIfTooLateForBooking(t *testing.T) {
	soon := time.Now().Add(30 * time.Minute)
	tr := &models.TrainingSchedule{
		TrainingDate: soon,
		StartTime:    soon,
	}
	if err := errIfTooLateForBooking(tr, "Запись"); err == nil {
		t.Fatal("expected error within 2h")
	}

	later := time.Now().Add(3 * time.Hour)
	tr2 := &models.TrainingSchedule{
		TrainingDate: later,
		StartTime:    later,
	}
	if err := errIfTooLateForBooking(tr2, "Запись"); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}
