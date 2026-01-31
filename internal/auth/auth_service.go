package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go-mini-erp/internal/dbgen"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserInactive       = errors.New("user is inactive")
	ErrUsernameExists     = errors.New("username already exists")
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
)

type Service interface {
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
	Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error)
	GetProfile(ctx context.Context, userID uuid.UUID) (*UserProfile, error)
	Logout(ctx context.Context, userID uuid.UUID) error
}

type service struct {
	repo      Repository
	dbgen     *sql.DB
	jwtSecret []byte
}

func NewService(repo Repository, database *sql.DB) Service {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your-secret-key-change-in-production"
	}

	return &service{
		repo:      repo,
		dbgen:     database,
		jwtSecret: []byte(secret),
	}
}

// Request/Response DTOs
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
}

type LoginResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	TokenType    string   `json:"token_type"`
	ExpiresIn    int      `json:"expires_in"`
	User         UserInfo `json:"user"`
}

type RegisterResponse struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	CreatedAt time.Time `json:"created_at"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

type UserInfo struct {
	ID       uuid.UUID  `json:"id"`
	Username string     `json:"username"`
	Email    string     `json:"email"`
	FullName string     `json:"full_name"`
	Roles    []RoleInfo `json:"roles"`
}

type RoleInfo struct {
	ID   uuid.UUID `json:"id"`
	Code string    `json:"code"`
	Name string    `json:"name"`
}

type UserProfile struct {
	ID          uuid.UUID  `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	FullName    string     `json:"full_name"`
	IsActive    bool       `json:"is_active"`
	LastLoginAt *time.Time `json:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at"`
	Roles       []RoleInfo `json:"roles"`
	Menus       []MenuInfo `json:"menus"`
}

type MenuInfo struct {
	ID        uuid.UUID  `json:"id"`
	ParentID  *uuid.UUID `json:"parent_id"`
	Code      string     `json:"code"`
	Name      string     `json:"name"`
	Path      *string    `json:"path"`
	Icon      *string    `json:"icon"`
	CanCreate bool       `json:"can_create"`
	CanRead   bool       `json:"can_read"`
	CanUpdate bool       `json:"can_update"`
	CanDelete bool       `json:"can_delete"`
}

// JWT Claims
type Claims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

// Login authenticates user and returns tokens
func (s *service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	// Get user by username
	user, err := s.repo.GetUserByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if user is active
	if !user.IsActive.Bool {
		return nil, ErrUserInactive
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Get user roles
	roles, err := s.repo.GetUserRoles(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	roleCodes := make([]string, len(roles))
	roleInfos := make([]RoleInfo, len(roles))
	for i, role := range roles {
		roleCodes[i] = role.Code
		roleInfos[i] = RoleInfo{
			ID:   role.ID,
			Code: role.Code,
			Name: role.Name,
		}
	}

	// Generate tokens
	accessToken, err := s.generateAccessToken(user.ID, user.Username, user.Email, roleCodes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Update last login
	if err := s.repo.UpdateUserLastLogin(ctx, user.ID); err != nil {
		// Log error but don't fail the login
		fmt.Printf("Failed to update last login: %v\n", err)
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    900, // 15 minutes
		User: UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			FullName: user.FullName,
			Roles:    roleInfos,
		},
	}, nil
}

// Register creates a new user
func (s *service) Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error) {
	// Check if username exists
	usernameExists, err := s.repo.CheckUsernameExists(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check username: %w", err)
	}
	if usernameExists {
		return nil, ErrUsernameExists
	}

	// Check if email exists
	emailExists, err := s.repo.CheckEmailExists(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}
	if emailExists {
		return nil, ErrEmailExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user, err := s.repo.CreateUser(ctx, dbgen.CreateUserParams{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FullName:     req.FullName,
		IsActive:     dbgen.NewNullBool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &RegisterResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FullName:  user.FullName,
		CreatedAt: user.CreatedAt.Time,
	}, nil
}

// RefreshToken generates new access token from refresh token
func (s *service) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	// Parse and validate refresh token
	token, err := jwt.ParseWithClaims(refreshToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Check if token is expired
	if claims.ExpiresAt.Before(time.Now()) {
		return nil, ErrTokenExpired
	}

	// Get user to ensure still active
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, ErrInvalidToken
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if !user.IsActive.Bool {
		return nil, ErrUserInactive
	}

	// Get roles
	roles, err := s.repo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	roleCodes := make([]string, len(roles))
	for i, role := range roles {
		roleCodes[i] = role.Code
	}

	// Generate new tokens
	newAccessToken, err := s.generateAccessToken(user.ID, user.Username, user.Email, roleCodes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := s.generateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    900, // 15 minutes
	}, nil
}

// GetProfile returns user profile with roles and menus
func (s *service) GetProfile(ctx context.Context, userID uuid.UUID) (*UserProfile, error) {
	// Get user
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Get roles
	roles, err := s.repo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	roleInfos := make([]RoleInfo, len(roles))
	for i, role := range roles {
		roleInfos[i] = RoleInfo{
			ID:   role.ID,
			Code: role.Code,
			Name: role.Name,
		}
	}

	// Get menus
	menus, err := s.repo.GetUserMenus(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user menus: %w", err)
	}

	menuInfos := make([]MenuInfo, len(menus))
	for i, menu := range menus {
		menuInfos[i] = MenuInfo{
			ID:        menu.ID,
			ParentID:  nullUUIDToPtr(menu.ParentID),
			Code:      menu.Code,
			Name:      menu.Name,
			Path:      nullStringToPtr(menu.Path),
			Icon:      nullStringToPtr(menu.Icon),
			CanCreate: menu.CanCreate.(int64) > 0,
			CanRead:   menu.CanRead.(int64) > 0,
			CanUpdate: menu.CanUpdate.(int64) > 0,
			CanDelete: menu.CanDelete.(int64) > 0,
		}
	}

	return &UserProfile{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		FullName:    user.FullName,
		IsActive:    user.IsActive.Bool,
		LastLoginAt: nullTimeToPtr(user.LastLoginAt),
		CreatedAt:   user.CreatedAt.Time,
		Roles:       roleInfos,
		Menus:       menuInfos,
	}, nil
}

// Logout invalidates user session (placeholder for token blacklist)
func (s *service) Logout(ctx context.Context, userID uuid.UUID) error {
	// TODO: Implement token blacklist using Redis
	// For now, client-side will remove token
	return nil
}

// Token generation helpers
func (s *service) generateAccessToken(userID uuid.UUID, username, email string, roles []string) (string, error) {
	claims := Claims{
		UserID:   userID.String(),
		Username: username,
		Email:    email,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *service) generateRefreshToken(userID uuid.UUID) (string, error) {
	claims := Claims{
		UserID: userID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// Helper functions
func nullStringToPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}

func nullUUIDToPtr(nu uuid.NullUUID) *uuid.UUID {
	if !nu.Valid {
		return nil
	}
	return &nu.UUID
}

func nullTimeToPtr(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}
	return &nt.Time
}
