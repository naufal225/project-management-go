package services

import (
	"github.com/naufal225/project-management-go/models"
	"github.com/naufal225/project-management-go/repositories"
)

type LabelService interface {
	Create(label *models.Label) error
	Update(label *models.Label) error
	Delete(id uint) error
	GetByPublicID(publicID string) (*models.Label, error)
}

type labelService struct {
	labelRepo repositories.LabelRepository
}

func NewLabelService(r repositories.LabelRepository) LabelService {
	return &labelService{labelRepo: r}
}

func (s *labelService) Create(label *models.Label) error {
	return s.labelRepo.Create(label)
}

func (s *labelService) Update(label *models.Label) error {
	return s.labelRepo.Update(label)
}

func (s *labelService) Delete(id uint) error {
	return s.labelRepo.Delete(id)
}

func (s *labelService) GetByPublicID(publicID string) (*models.Label, error) {
	return s.labelRepo.FindByPublicID(publicID)
}


