package repositories

import (
	"github.com/naufal225/project-management-go/config"
	"github.com/naufal225/project-management-go/models"
)

type LabelRepository interface {
	Create(label *models.Label) error
	Update(label *models.Label) error
	Delete(id uint) error
	FindByPublicID(publicID string) (*models.Label, error)
}

type labelRepository struct {
}

func NewLabelRepository() LabelRepository {
	return &labelRepository{}
}

func (r *labelRepository) Create(label *models.Label) error {
	return config.DB.Create(label).Error
}

func (r *labelRepository) Update(label *models.Label) error {
	return config.DB.Save(label).Error
}

func (r *labelRepository) Delete(id uint) error {
	return config.DB.Delete(&models.Label{}, id).Error
}

func (r *labelRepository) FindByPublicID(publicID string) (*models.Label, error) {
	var label models.Label
	err := config.DB.Where("public_id = ?", publicID).First(&label).Error
	return &label, err
}
