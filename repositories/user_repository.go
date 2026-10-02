package repositories

import (
	"strings"

	"github.com/google/uuid"
	"github.com/naufal225/project-management-go/config"
	"github.com/naufal225/project-management-go/models"
)

type UserRepository interface {
	Create(user *models.User) error
	FindByEmail(email string) (*models.User, error)
	FindByID(id uint) (*models.User, error)
	FindByPublicID(publicID string) (*models.User, error)
	FindAllPagination(filter, sort string, limit, offset int) ([]models.User, int64, error)
	Update(user *models.User) error
	Delete(id uuid.UUID) error
} 

type userRepository struct {

}

func NewUserRepository() UserRepository {
	return &userRepository{}
}

func (r *userRepository) Create(user *models.User) error {
	return config.DB.Create(user).Error
}

func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := config.DB.Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *userRepository) FindByID (id uint) (*models.User, error) {
	var user models.User
	err := config.DB.First(&user, id).Error
	return &user, err
}

func (r *userRepository) FindByPublicID (publicID string) (*models.User, error) {
	var user models.User
	err := config.DB.Where("public_id = ?", publicID).First(&user).Error
	return &user, err
}

func (r *userRepository) FindAllPagination(filter, sort string, limit, offset int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	db := config.DB.Model(&models.User{})
	
	if filter != "" {
		filterPattern := "%" + filter + "%"
		db.Where("name ILIKE ? OR email ILIKE ?", filterPattern, filterPattern)
	}

	if sort != "" {
		if sort == "-id" {
			sort = "-internal_id"
		} else if sort == "id" {
			sort = "internal_id"
		}
	
		if strings.HasPrefix(sort, "-") {
			sort = strings.TrimPrefix(sort, "-") + " DESC"
		} else {
			sort += " ASC"
		}

		db = db.Order(sort)
	}

	err := db.Limit(limit).Offset(offset).Find(&users).Error

	if err := db.Count(&total).Error; err != nil {
		return  nil, 0, err
	}

	return users, total, err
}

func (r *userRepository) Update(user *models.User) error {
	return config.DB.Model(&models.User{}).
		Where("public_id = ?", user.PublicID).Updates(map[string]interface{}{
		"name":user.Name,
		"email":user.Email,
		"role":user.Role,
	}).Error
}

func (r *userRepository) Delete(id uuid.UUID) error {
	return config.DB.Where("public_id = ?", id).
		Delete(&models.User{}).Error
}