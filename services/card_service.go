package services

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/naufal225/project-management-go/config"
	"github.com/naufal225/project-management-go/models"
	"github.com/naufal225/project-management-go/models/types"
	"github.com/naufal225/project-management-go/repositories"
	"gorm.io/gorm"
)

type CardService interface {
	Create(card *models.Card, listPublicID string) error
	Update(card *models.Card, listPublicID string) error
	Delete(id uint) error

	GetByListID(listPublicID string) ([]models.Card, error)
	GetByID(id uint) (*models.Card, error)
	GetByPublicID(publicID string) (*models.Card, error)

	AddLabel(cardPublicID, labelPublicID string) error
	RemoveLabel(cardPublicID, labelPublicID string) error

	AddAssignees(cardID uint, userID []uint) error
	RemoveAssignees(cardID uint, userID uint) error
}

type cardService struct {
	cardRepo  repositories.CardRepository
	listRepo  repositories.ListRepository
	userRepo  repositories.UserRepository
	labelRepo repositories.LabelRepository
}

func NewCardService(
	cardRepo repositories.CardRepository,
	listRepo repositories.ListRepository,
	userRepo repositories.UserRepository,
	labelRepo repositories.LabelRepository,
) CardService {
	return &cardService{
		cardRepo: cardRepo, listRepo: listRepo, userRepo: userRepo, labelRepo: labelRepo,
	}
}

func (s *cardService) Create(card *models.Card, listPublicID string) error {
	list, err := s.listRepo.FindByPublicID(listPublicID)
	if err != nil {
		return fmt.Errorf("list not found: %w", err)
	}

	card.ListID = list.InternalID

	if card.PublicID == uuid.Nil {
		card.PublicID = uuid.New()
	}
	card.CreatedAt = time.Now()

	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.Create(card).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create card: %w", err)
	}

	var position models.CardPosition
	if err := tx.Model(&models.CardPosition{}).
		Where("list_internal_id = ?", list.InternalID).
		First(&position).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			position = models.CardPosition{
				PublicID:  uuid.New(),
				ListID:    list.InternalID,
				CardOrder: types.UUIDArray{card.PublicID},
			}

			if err := tx.Create(&position).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create position: %w", err)
			}
		} else {
			tx.Rollback()
			return fmt.Errorf("failed to get card position: %w", err)
		}
	} else {
		position.CardOrder = append(position.CardOrder, card.PublicID)
		if err := tx.Model(&models.CardPosition{}).
			Where("internal_id = ?", position.InternalID).
			Update("card_order", position.CardOrder).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update card position: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("transaction commit failed: %w", err)
	}

	return nil
}

func (s *cardService) Update(card *models.Card, listPublicID string) error {
	exisitingCard, err := s.cardRepo.FindByPublicID(card.PublicID.String())
	if err != nil {
		return fmt.Errorf("card not found: %w", err)
	}

	newList, err := s.listRepo.FindByPublicID(listPublicID)
	if err != nil {
		return fmt.Errorf("list not found: %w", err)
	}

	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// jika pindah list -> hapus dari posisi list lama dan tambah ke list baru

	if exisitingCard.ListID != newList.InternalID {
		// hapus dari list lama
		var oldPos models.CardPosition
		if err := tx.Where("list_internal_id = ?", exisitingCard.ListID).First(&oldPos).Error; err != nil {
			filtered := make(types.UUIDArray, 0, len(oldPos.CardOrder))
			for _, id := range oldPos.CardOrder {
				if id != exisitingCard.PublicID {
					filtered = append(filtered, id)
				}
			}
			// update
			if err := tx.Model(&models.CardPosition{}).Where("internal_id = ?", oldPos.InternalID).
				Update("card_order", types.UUIDArray(filtered)).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update old card position: %w", err)
			}

		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return fmt.Errorf("failed to get old card position: %w", err)
		}

		// tambah ke list baru

		var newPos models.CardPosition
		res := tx.Where("list_internal_id = ?", newList.InternalID).First(&newPos)
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			newPos = models.CardPosition{
				PublicID:  uuid.New(),
				ListID:    newList.InternalID,
				CardOrder: types.UUIDArray{exisitingCard.PublicID},
			}
			if err := tx.Create(&newPos).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create card position for new list: %w", err)
			}
		} else if res.Error == nil {
			updateOrder := append(newPos.CardOrder, exisitingCard.PublicID)
			if err := tx.Model(&models.CardPosition{}).Where("internal_id = ?", newPos.InternalID).
				Update("card_order", types.UUIDArray(updateOrder)).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update card position: %w", err)
			}
		} else {
			tx.Rollback()
			return fmt.Errorf("failed to get new card position: %w", res.Error)
		}
	}

	// update data card

	card.InternalID = exisitingCard.InternalID
	card.PublicID = exisitingCard.PublicID
	card.ListID = exisitingCard.ListID

	if err := tx.Save(card).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update card: %w", err)
	}

	// commit

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("transaction commit failed: %w", err)
	}

	return nil
}

func (s *cardService) Delete(id uint) error {
	return s.cardRepo.Delete(id)
}

func (s *cardService) GetByListID(listPublicID string) ([]models.Card, error) {
	// verifikasi listnya ada
	list, err := s.listRepo.FindByPublicID(listPublicID)
	if err != nil {
		return nil, fmt.Errorf("list not found: %w", err)
	}

	// ambil card position
	position, err := s.cardRepo.FindCardPositionByListID(list.InternalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get card position: %w", err)
	}

	// ambil semua card di list

	cards, err := s.cardRepo.FindByListID(listPublicID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cards: %w", err)
	}

	// sorting
	if position != nil && len(position.CardOrder) > 0 {
		cards = sortCardByPosition(cards, position.CardOrder)
	}

	return cards, nil
}

func sortCardByPosition(cards []models.Card, order []uuid.UUID) []models.Card {
	orderMap := make(map[uuid.UUID]int)
	for i, id := range order {
		orderMap[id] = i
	}

	defaultIndex := len(order)

	// sorting slice
	sort.SliceStable(cards, func(i, j int) bool {
		idxI, okI := orderMap[cards[i].PublicID]
		if !okI {
			idxI = defaultIndex
		}

		idxJ, okJ := orderMap[cards[j].PublicID]
		if !okJ {
			idxJ = defaultIndex
		}

		if idxI == idxJ {
			return cards[i].CreatedAt.Before(cards[j].CreatedAt)
		}

		return idxI < idxJ
	})

	return cards
}

func (s *cardService) GetByID(id uint) (*models.Card, error) {
	return s.cardRepo.FindByID(id)
}

func (s *cardService) GetByPublicID(publicID string) (*models.Card, error) {
	return s.cardRepo.FindByPublicID(publicID)
}

func (s *cardService) AddLabel(cardPublicID, labelPublicID string) error {
	card, err := s.cardRepo.FindByPublicID(cardPublicID)
	if err != nil {
		return errors.New("card not found")
	}

	label, err := s.labelRepo.FindByPublicID(labelPublicID)
	if err != nil {
		return errors.New("label not found")
	}

	return s.cardRepo.AddLabel(uint(card.InternalID), uint(label.InternalID))
}

func (s *cardService) RemoveLabel(cardPublicID, labelPublicID string) error {
	card, err := s.cardRepo.FindByPublicID(cardPublicID)
	if err != nil {
		return errors.New("label not found")
	}

	label, err := s.labelRepo.FindByPublicID(labelPublicID)
	if err != nil {
		return errors.New("label not found")
	}

	return s.cardRepo.RemoveLabel(uint(card.InternalID), uint(label.InternalID))
}

func (s *cardService) AddAssignees(cardID uint, userID []uint) error {
	_, err := s.cardRepo.FindByID(cardID)
	if err != nil {
		return errors.New("card not found")
	}

	var usersToAssign []uint
	for _, uID := range userID {

		if _, err := s.userRepo.FindByID(uID); err != nil {
			return fmt.Errorf("user not found: %w", err)
		}

		// Masih harus mengecek
		// apakah user sudah di assign?
		count, err := s.cardRepo.IsUserAssignedToCard(uID, cardID)

		if err != nil {
			return fmt.Errorf("error while counting: %w", err)
		}

		if count == 0 {
			usersToAssign = append(usersToAssign, uID)
		}
	}

	return s.cardRepo.AddAssignees(cardID, usersToAssign)
}

func (s *cardService) RemoveAssignees(cardID uint, userID uint) error {
	if _, err := s.cardRepo.FindByID(cardID); err != nil {
		return errors.New("card not found")
	}

	if _, err := s.userRepo.FindByID(userID); err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Masih harus mengecek
	// apakah user sudah di assign?
	count, err := s.cardRepo.IsUserAssignedToCard(userID, cardID)

	if err != nil {
		return fmt.Errorf("error while counting: %w", err)
	}

	if count == 0 {
		return errors.New("user is not assigned")
	}

	return s.cardRepo.RemoveAssignees(cardID, userID)
}
