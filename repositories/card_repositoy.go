package repositories

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/naufal225/project-management-go/config"
	"github.com/naufal225/project-management-go/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CardRepository interface {
	Create(card *models.Card) error
	Update(card *models.Card) error
	Delete(id uint) error
	FindByID(id uint) (*models.Card, error)
	FindByPublicID(publicID string) (*models.Card, error)
	FindByListID(listID string) ([]models.Card, error)

	FindCardPositionByListID(id int64) (*models.CardPosition, error)
	UpdatePosition(listID string, position []string) error

	AddLabel(cardID, labelID uint) error
	RemoveLabel(cardID, labelID uint) error

	AddAssignees(cardID uint, userID []uint) error
	RemoveAssignees(cardID uint, userID uint) error
	FindAssigneesByUserID(userID uint) ([]models.CardAssignee, error)
	IsUserAssignedToCard(userID uint, cardID uint) (int64, error)
}

type cardRepository struct {
}

func NewCardRepository() CardRepository {
	return &cardRepository{}
}

func (r *cardRepository) Create(card *models.Card) error {
	return config.DB.Create(card).Error
}

func (r *cardRepository) Update(card *models.Card) error {
	return config.DB.Save(card).Error
}

func (r *cardRepository) Delete(id uint) error {
	return config.DB.Delete(&models.Card{}, id).Error
}

func (r *cardRepository) FindByID(id uint) (*models.Card, error) {
	var card models.Card
	err := config.DB.Preload("Labels").Preload("Assignees").First(&card, id).Error
	return &card, err
}

func (r *cardRepository) FindByPublicID(publicID string) (*models.Card, error) {
	var card models.Card
	if err := config.DB.Preload("Assignees.User", func(tx *gorm.DB) *gorm.DB {
		return tx.Select("internal_id", "public_id", "name", "email")
	}).Preload("Attachments").Where("public_id = ?", publicID).First(&card).Error; err != nil {
		return nil, err
	}

	baseUrl := config.AppConfig.AppURL

	for i := range card.Attachments {
		card.Attachments[i].FileURL = fmt.Sprintf("%s/files/%s",
			baseUrl,
			filepath.Base(card.Attachments[i].File),
		)
	}

	return &card, nil
}

func (r *cardRepository) FindByListID(listID string) ([]models.Card, error) {
	var cards []models.Card
	err := config.DB.Joins("JOIN lists ON lists.internal_id = cards.list_internal_id").
		Where("lists.public_id = ?", listID).
		Order("position ASC").
		Find(&cards).Error
	return cards, err
}

func (r *cardRepository) FindCardPositionByListID(id int64) (*models.CardPosition, error) {
	var position models.CardPosition
	err := config.DB.Where("list_internal_id = ?", id).First(&position).Error
	if err != nil {
		return nil, err
	}

	return &position, err
}

func (r *cardRepository) UpdatePosition(listID string, position []string) error {
	return config.DB.Model(&models.CardPosition{}).Where("list_internal_id = (SELECT internal_id FROM lists where public_id = ?)", listID).Update("card_order", position).Error

}

func (r *cardRepository) AddLabel(cardID, labelID uint) error {
	cardLabel := models.CardLabel{
		CardID:  int64(cardID),
		LabelID: int64(labelID),
	}

	return config.DB.Create(&cardLabel).Error
}

func (r *cardRepository) RemoveLabel(cardID, labelID uint) error {
	return config.DB.Where("card_internal_id = ? AND label_internal_id = ?", cardID, labelID).Delete(&models.CardLabel{}).Error
}

func (r *cardRepository) AddAssignees(cardID uint, userID []uint) error {
	var assignees []models.CardAssignee
	for _, uid := range userID {
		assignees = append(assignees, models.CardAssignee{
			CardID: int64(cardID),
			UserID: int64(uid),
		})
	}

	return config.DB.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(&assignees, 100).Error
}

func (r *cardRepository) RemoveAssignees(cardID uint, userID uint) error {
	return config.DB.Where("card_internal_id = ? AND user_internal_id = ?", cardID, userID).Delete(&models.CardAssignee{}).Error
}

func (c *cardRepository) FindAssigneesByUserID(userID uint) ([]models.CardAssignee, error) {
	var assignees []models.CardAssignee

	err := config.DB.Joins("JOIN users on card_assignees.user_internal_id = users.internal_id").
		Where("users.internal_id = ?", userID).
		Find(&assignees).Error

	return assignees, err
}

func (c *cardRepository) IsUserAssignedToCard(userID uint, cardID uint) (int64, error) {
	var count int64
	err := config.DB.Model(&models.CardAssignee{}).
		Where("user_internal_id = ? AND card_internal_id = ?", userID, cardID).Count(&count).Error
	
	if err != nil {
		return 0, errors.New("error counting")
	}

	return count, nil
		
} 