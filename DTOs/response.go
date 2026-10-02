package dtos

import (
	"time"

	"github.com/google/uuid"
)

type BoardResponse struct {
	PublicID      uuid.UUID   `json:"public_id" db:"public_id"`
	Title         string     `json:"title" db:"title"`
	Description   string     `json:"description" db:"description"`
	OwnerID       int64      `json:"owner_internal_id" db:"owner_intenal_id" gorm:"column:owner_internal_id"`
	OwnerPublicID uuid.UUID  `json:"owner_public_id" db:"owner_public_id"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	DueDate       *time.Time `json:"due_date,omitempty" db:"due_date"`
}