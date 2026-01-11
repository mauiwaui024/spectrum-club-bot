package group_serivce

import (
	"spectrum-club-bot/internal/models"
	"spectrum-club-bot/internal/repository"
	"spectrum-club-bot/internal/service"
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
