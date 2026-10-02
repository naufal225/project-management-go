package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/naufal225/project-management-go/models"
	"github.com/naufal225/project-management-go/repositories"
)

type AttachmentService interface {
	GetByPublicID(publicID uuid.UUID) (*models.CardAttachment, error)
	Create(cardPublicID, userPublicID, fileName string) (*models.CardAttachment, error) 
	DeleteByPublicID(publicID uuid.UUID) error
	GetByCardID(cardPublicID string) ([]models.CardAttachment, error)
}

type attachmentService struct {
	attachmentRepo repositories.AttachmentRepository
	cardRepo repositories.CardRepository
	userRepo repositories.UserRepository
}

func NewAttachmentService(
	attachmentRepo repositories.AttachmentRepository,
	cardRepo repositories.CardRepository,
	userRepo repositories.UserRepository,
) AttachmentService {
	return &attachmentService{
		attachmentRepo: attachmentRepo,
		cardRepo: cardRepo,
		userRepo: userRepo,
	}
}

func (s *attachmentService) GetByPublicID(publicID uuid.UUID) (*models.CardAttachment, error) {
	return s.attachmentRepo.GetByPublicID(publicID)
}

func (s *attachmentService) Create(cardPublicID, userPublicID, fileName string) (*models.CardAttachment, error) {
	card, err := s.cardRepo.FindByPublicID(cardPublicID)
	if err != nil {
		return nil, errors.New("card not found")
	}

	user, err := s.userRepo.FindByPublicID(userPublicID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	attach := &models.CardAttachment{
		PublicID: uuid.New(),
		CardID: card.InternalID,
		UserID: user.InternalID,
		File: fileName,
		CreatedAt: time.Now(),
	}

	if err := s.attachmentRepo.Create(attach); err != nil {
		return nil, err
	}

	return attach, nil
}

func (s *attachmentService) DeleteByPublicID(publicID uuid.UUID) error {
	return s.attachmentRepo.DeleteByPublicID(publicID)
}

func (s *attachmentService) GetByCardID(cardPublicID string) ([]models.CardAttachment, error) {
	_, err := s.cardRepo.FindByPublicID(cardPublicID)
	if err != nil {
		return nil, errors.New("card not found")
	}

	return s.attachmentRepo.FindByCardID(cardPublicID)
}