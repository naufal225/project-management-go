package dtos

import (
	"time"

	"github.com/google/uuid"
)

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type AdminUpdateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type CreateBoardRequest struct {
	Title         string     `json:"title" db:"title"`
	Description   string     `json:"description" db:"description"`
	OwnerPublicID uuid.UUID  `json:"owner_public_id" db:"owner_public_id"`
	DueDate       *time.Time `json:"due_date,omitempty" db:"due_date"`
}

type UpdateBoardRequest struct {
	Title       string     `json:"title" db:"title"`
	Description string     `json:"description" db:"description"`
	DueDate     *time.Time `json:"due_date,omitempty" db:"due_date"`
}

type CreateCardRequest struct {
	ListPublicID string    `json:"list_id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	DueDate      time.Time `json:"due_date"`
	Position     int       `json:"position"`
}

type UpdateCardRequest struct {
	ListPublicID string     `json:"list_id"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	DueDate      *time.Time `json:"due_date"`
	Position     int        `json:"position"`
}
