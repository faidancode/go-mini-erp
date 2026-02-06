package role

import (
	"time"

	"github.com/google/uuid"
)

type CreateRoleRequest struct {
	Code        string  `json:"code" binding:"required,min=3,max=50"`
	Name        string  `json:"name" binding:"required,min=3,max=100"`
	Description *string `json:"description"`
}

type UpdateRoleRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=100"`
	Description string `json:"description"`
	IsActive    *bool  `json:"isActive" binding:"required"`
}

type RoleResponse struct {
	ID          uuid.UUID `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	IsActive    bool      `json:"isActive"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type RoleProfile struct {
	ID          uuid.UUID
	Code        string
	Name        string
	Description string
	IsActive    bool
}
