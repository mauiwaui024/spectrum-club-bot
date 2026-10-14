package plan

import (
	"spectrum-club-bot/internal/models"
	planrepo "spectrum-club-bot/internal/repository/plan"
	"spectrum-club-bot/internal/service"
)

type planService struct {
	repo *planrepo.Repository
}

func NewPlanService(repo *planrepo.Repository) service.PlanService {
	return &planService{repo: repo}
}

func (s *planService) ListTemplates() ([]models.PlanTemplate, error) {
	return s.repo.ListTemplates()
}

func (s *planService) GetTemplate(id int) (*models.PlanTemplate, error) {
	return s.repo.GetTemplate(id)
}

func (s *planService) CreateTemplate(title, description string, items []models.PlanItem) (*models.PlanTemplate, error) {
	return s.repo.CreateTemplate(title, description, items)
}

func (s *planService) UpdateTemplate(id int, title, description string, items []models.PlanItem) error {
	return s.repo.UpdateTemplate(id, title, description, items)
}

func (s *planService) DeleteTemplate(id int) error {
	return s.repo.DeleteTemplate(id)
}

func (s *planService) GetTrainingPlan(trainingID int) (*models.TrainingPlan, error) {
	return s.repo.GetTrainingPlan(trainingID)
}

func (s *planService) SaveTrainingPlan(trainingID int, sourceTemplateID *int, description string, items []models.PlanItem) (*models.TrainingPlan, error) {
	return s.repo.SaveTrainingPlan(trainingID, sourceTemplateID, description, items)
}

func (s *planService) ListExercises() ([]models.Exercise, error) {
	return s.repo.ListExercises()
}

func (s *planService) GetExercise(id int) (*models.Exercise, error) {
	return s.repo.GetExercise(id)
}

func (s *planService) CreateExercise(name, description string) (*models.Exercise, error) {
	return s.repo.CreateExercise(name, description)
}

func (s *planService) UpdateExercise(id int, name, description string) error {
	return s.repo.UpdateExercise(id, name, description)
}

func (s *planService) DeleteExercise(id int) error {
	return s.repo.DeleteExercise(id)
}
