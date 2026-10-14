package group_serivce

import (
	"fmt"
	"spectrum-club-bot/internal/models"
	"spectrum-club-bot/internal/repository"
	"spectrum-club-bot/internal/service"
	"strings"
)

type trainingGroupService struct {
	groupRepo repository.TrainingGroupRepository
}

func NewTrainingGroupService(groupRepo repository.TrainingGroupRepository) service.TrainingGroupService {
	return &trainingGroupService{
		groupRepo: groupRepo,
	}
}
func (s *trainingGroupService) GetAllGroups() ([]models.TrainingGroup, error) {
	return s.groupRepo.GetAllGroups()
}

func (s *trainingGroupService) GetGroupByID(id int) (*models.TrainingGroup, error) {
	return s.groupRepo.GetGroupByID(id)
}

func (s *trainingGroupService) GetGroupsForAge(age int) ([]models.TrainingGroup, error) {
	groups, err := s.groupRepo.GetAllGroups()
	if err != nil {
		return nil, err
	}

	var matchingGroups []models.TrainingGroup
	for _, group := range groups {
		if age >= group.AgeMin && (group.AgeMax == nil || age <= *group.AgeMax) {
			matchingGroups = append(matchingGroups, group)
		}
	}

	return matchingGroups, nil
}

func (s *trainingGroupService) CreateGroup(group *models.TrainingGroup) error {
	group.Name = strings.TrimSpace(group.Name)
	group.Code = strings.TrimSpace(group.Code)
	if group.Name == "" || group.Code == "" {
		return fmt.Errorf("name and code required")
	}
	return s.groupRepo.Create(group)
}

func (s *trainingGroupService) UpdateGroup(id int, updates map[string]interface{}) error {
	existing, err := s.groupRepo.GetGroupByID(id)
	if err != nil || existing == nil {
		return fmt.Errorf("group not found")
	}
	name := existing.Name
	code := existing.Code
	desc := existing.Description
	ageMin := existing.AgeMin
	ageMax := existing.AgeMax
	if v, ok := updates["name"].(string); ok && strings.TrimSpace(v) != "" {
		name = strings.TrimSpace(v)
	}
	if v, ok := updates["code"].(string); ok && strings.TrimSpace(v) != "" {
		code = strings.TrimSpace(v)
	}
	if v, ok := updates["description"].(string); ok {
		desc = v
	}
	if v, ok := updates["age_min"].(float64); ok {
		ageMin = int(v)
	} else if v, ok := updates["age_min"].(int); ok {
		ageMin = v
	}
	if v, ok := updates["age_max"].(float64); ok {
		n := int(v)
		ageMax = &n
	} else if v, ok := updates["age_max"].(int); ok {
		ageMax = &v
	} else if updates["age_max"] == nil {
		if _, present := updates["age_max"]; present {
			ageMax = nil
		}
	}
	return s.groupRepo.Update(id, name, code, desc, ageMin, ageMax)
}

func (s *trainingGroupService) DeleteGroup(id int) error {
	n, err := s.groupRepo.CountFutureTrainings(id)
	if err != nil {
		return err
	}
	if n > 0 {
		return fmt.Errorf("нельзя удалить группу: есть %d будущих тренировок", n)
	}
	return s.groupRepo.Delete(id)
}
