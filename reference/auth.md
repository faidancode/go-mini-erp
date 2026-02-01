## DTO
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

## Repo
package auth

import (
	"context"

	db "go-mini-erp/internal/shared/database/sqlc"

	"github.com/google/uuid"
)

//go:generate mockgen -source=auth_repo.go -destination=mocks/auth_repository_mock.go -package=mocks

// Repository defines auth data access contract
type Repository interface {
	GetUserByUsername(ctx context.Context, username string) (db.GetUserByUsernameRow, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (db.GetUserByIDRow, error)
	GetUserByEmail(ctx context.Context, email string) (db.GetUserByEmailRow, error)

	CreateUser(ctx context.Context, arg db.CreateUserParams) (db.CreateUserRow, error)
	UpdateUserLastLogin(ctx context.Context, id uuid.UUID) error

	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]db.GetUserRolesRow, error)
	GetUserMenus(ctx context.Context, userID uuid.UUID) ([]db.GetUserMenusRow, error)

	AssignRoleToUser(ctx context.Context, arg db.AssignRoleToUserParams) (db.AssignRoleToUserRow, error)
	RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error

	CheckUsernameExists(ctx context.Context, username string) (bool, error)
	CheckEmailExists(ctx context.Context, email string) (bool, error)
}

// repository is concrete implementation
// depends on sqlc interface, NOT concrete Queries
type repository struct {
	q db.Querier
}

// NewRepository creates auth repository
func NewRepository(q db.Querier) Repository {
	return &repository{
		q: q,
	}
}

// ==========================
// User
// ==========================

func (r *repository) GetUserByUsername(
	ctx context.Context,
	username string,
) (db.GetUserByUsernameRow, error) {
	return r.q.GetUserByUsername(ctx, username)
}

## Service
package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"

	dbgen "go-mini-erp/internal/shared/database/sqlc"
	"go-mini-erp/internal/shared/util/dbutil"
)

//go:generate mockgen -source=auth_service.go -destination=mocks/auth_service_mock.go -package=mocks
type Service interface {
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
	Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error)
	GetProfile(ctx context.Context, userID uuid.UUID) (*UserProfile, error)
	Logout(ctx context.Context, userID uuid.UUID) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]RoleInfo, error)
	AssignRoleToUser(ctx context.Context, userID, roleID, assignedBy uuid.UUID) (*RoleAssignmentResponse, error)
	RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error
}

type service struct {
	repo       Repository
	queries    *dbgen.Queries
	jwtManager JWTManager
}

func NewService(
	repo Repository,
	queries *dbgen.Queries,
	jwtManager JWTManager,
) Service {
	return &service{
		repo:       repo,
		queries:    queries,
		jwtManager: jwtManager,
	}
}

// ----------------------------------------------------
// Login
// ----------------------------------------------------

func (s *service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// IsActive sekarang *bool, jadi aman
	if user.IsActive == nil || !*user.IsActive {
		return nil, ErrUserInactive
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(req.Password),
	); err != nil {
		return nil, ErrInvalidCredentials
	}

	roles, err := s.repo.GetUserRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	roleCodes := make([]string, 0, len(roles))
	roleInfos := make([]RoleInfo, 0, len(roles))

	for _, r := range roles {
		roleCodes = append(roleCodes, r.Code)
		roleInfos = append(roleInfos, RoleInfo{
			ID:   r.ID,
			Code: r.Code,
			Name: r.Name,
		})
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(
		user.ID,
		user.Username,
		user.Email,
		roleCodes,
	)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	_ = s.repo.UpdateUserLastLogin(ctx, user.ID)

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    900,
		User: UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			FullName: user.FullName,
			Roles:    roleInfos,
		},
	}, nil
}

## Service Test
package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"

	"go-mini-erp/internal/auth"
	"go-mini-erp/internal/auth/mocks"
	db "go-mini-erp/internal/shared/database/sqlc"
	"go-mini-erp/internal/shared/util/dbutil"
)

/*
JWTManager stub:
- Tidak pakai gomock (lebih simpel)
- Fokus test business logic service
*/
type jwtManagerStub struct{}

func (j *jwtManagerStub) GenerateAccessToken(
	userID uuid.UUID,
	username, email string,
	roles []string,
) (string, error) {
	return "access-token", nil
}

func (j *jwtManagerStub) GenerateRefreshToken(userID uuid.UUID) (string, error) {
	return "refresh-token", nil
}

func (j *jwtManagerStub) ParseRefreshToken(token string) (*auth.Claims, error) {
	return nil, errors.New("not implemented")
}

// =======================
// LOGIN
// =======================

func TestLogin_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jwtStub := &jwtManagerStub{}
	repo := mocks.NewMockRepository(ctrl)
	service := auth.NewService(repo, nil, jwtStub)

	ctx := context.Background()
	userID := uuid.New()
	hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	repo.EXPECT().
		GetUserByUsername(ctx, "testuser").
		Return(db.GetUserByUsernameRow{
			ID:           userID,
			Username:     "testuser",
			Email:        "test@example.com",
			PasswordHash: string(hashed),
			FullName:     "Test User",
			IsActive:     dbutil.BoolPtr(true),
		}, nil)

	repo.EXPECT().
		GetUserRoles(ctx, userID).
		Return([]db.GetUserRolesRow{
			{ID: uuid.New(), Code: "admin", Name: "Administrator"},
		}, nil)

	repo.EXPECT().
		UpdateUserLastLogin(ctx, userID).
		Return(nil)

	result, err := service.Login(ctx, auth.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Bearer", result.TokenType)
	assert.Equal(t, 900, result.ExpiresIn)
	assert.Equal(t, "testuser", result.User.Username)
	assert.Len(t, result.User.Roles, 1)
	assert.Equal(t, "admin", result.User.Roles[0].Code)
}

## Handler
package auth

import (
	"errors"
	"go-mini-erp/internal/shared/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	{
		auth.POST("/login", h.Login)
		auth.POST("/register", h.Register)
		auth.POST("/refresh", h.RefreshToken)
		auth.POST("/logout", middleware.AuthMiddleware(), h.Logout)
		auth.GET("/profile", middleware.AuthMiddleware(), h.GetProfile)
	}
}

// Login godoc
// @Summary User login
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse
// @Failure 401 {object} map[string]string
// @Router /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	// Set refresh token as httpOnly cookie
	c.SetCookie(
		"refresh_token",
		result.RefreshToken,
		7*24*60*60, // 7 days
		"/",
		"",
		false, // Set to true in production with HTTPS
		true,  // httpOnly
	)

	c.JSON(http.StatusOK, result)
}

## Handler Test
package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"go-mini-erp/internal/auth"
	"go-mini-erp/internal/auth/mocks"
)

// Test Login - Success
func TestLoginHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	handler := auth.NewHandler(mockService)

	router := gin.Default()
	router.POST("/auth/login", handler.Login)

	userID := uuid.New()
	expectedResponse := &auth.LoginResponse{
		AccessToken:  "mock-access-token",
		RefreshToken: "mock-refresh-token",
		TokenType:    "Bearer",
		ExpiresIn:    900,
		User: auth.UserInfo{
			ID:       userID,
			Username: "testuser",
			Email:    "test@example.com",
			FullName: "Test User",
			Roles: []auth.RoleInfo{
				{
					ID:   uuid.New(),
					Code: "admin",
					Name: "Administrator",
				},
			},
		},
	}

	mockService.EXPECT().
		Login(gomock.Any(), auth.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}).
		Return(expectedResponse, nil).
		Times(1)

	// Create request
	body := map[string]string{
		"username": "testuser",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response auth.LoginResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "testuser", response.User.Username)
	assert.NotEmpty(t, response.AccessToken)
}