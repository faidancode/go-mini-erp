package auth

import (
	"time"

	"github.com/google/uuid"
)

// Request/Response DTOs
type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"fullName" binding:"required"`
}

type LoginResponse struct {
	AccessToken  string   `json:"accessToken"`
	RefreshToken string   `json:"refreshToken"`
	TokenType    string   `json:"tokenType"`
	ExpiresIn    int      `json:"expiresIn"`
	User         UserInfo `json:"user"`
}

type RegisterResponse struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	FullName  string    `json:"fullName"`
	CreatedAt time.Time `json:"createdAt"`
}

type TokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	TokenType    string `json:"tokenType"`
	ExpiresIn    int    `json:"expiresIn"`
}

type UserInfo struct {
	ID       uuid.UUID  `json:"id"`
	Username string     `json:"username"`
	Email    string     `json:"email"`
	FullName string     `json:"fullName"`
	Roles    []RoleInfo `json:"roles"`
}

type RoleInfo struct {
	ID   uuid.UUID `json:"id"`
	Code string    `json:"code"`
	Name string    `json:"name"`
}

type RoleAssignmentResponse struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"userId"`
	RoleID     uuid.UUID `json:"roleId"`
	AssignedAt time.Time `json:"assignedAt"`
}

type UserProfile struct {
	ID          uuid.UUID  `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	FullName    string     `json:"fullName"`
	IsActive    bool       `json:"isActive"`
	LastLoginAt *time.Time `json:"lastLoginAt"`
	CreatedAt   time.Time  `json:"createdAt"`
	Roles       []RoleInfo `json:"roles"`
	Menus       []MenuInfo `json:"menus"`
}

type MenuInfo struct {
	ID        uuid.UUID  `json:"id"`
	ParentID  *uuid.UUID `json:"parentId"`
	Code      string     `json:"code"`
	Name      string     `json:"name"`
	Path      *string    `json:"path"`
	Icon      *string    `json:"icon"`
	CanCreate bool       `json:"canCreate"`
	CanRead   bool       `json:"canRead"`
	CanUpdate bool       `json:"canUpdate"`
	CanDelete bool       `json:"canDelete"`
}
