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

func (s *service) Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error) {
	if exists, _ := s.repo.CheckUsernameExists(ctx, req.Username); exists {
		return nil, ErrUsernameExists
	}

	if exists, _ := s.repo.CheckEmailExists(ctx, req.Email); exists {
		return nil, ErrEmailExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return nil, fmt.Errorf("hash password failed: %w", err)
	}

	active := true

	user, err := s.repo.CreateUser(ctx, dbgen.CreateUserParams{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FullName:     req.FullName,
		IsActive:     &active, // FIX: *bool
	})
	if err != nil {
		return nil, fmt.Errorf("create user failed: %w", err)
	}

	return &RegisterResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FullName:  user.FullName,
		CreatedAt: user.CreatedAt.Time,
	}, nil
}

func (s *service) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]RoleInfo, error) {
	rolesRows, err := s.repo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	roles := make([]RoleInfo, len(rolesRows))
	for i, r := range rolesRows {
		roles[i] = RoleInfo{
			ID:   r.ID,
			Code: r.Code,
			Name: r.Name,
		}
	}
	return roles, nil
}

func (s *service) AssignRoleToUser(ctx context.Context, userID, roleID, assignedBy uuid.UUID) (*RoleAssignmentResponse, error) {
	// cek apakah user exist
	_, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	res, err := s.repo.AssignRoleToUser(ctx, dbgen.AssignRoleToUserParams{
		UserID:     userID,
		RoleID:     roleID,
		AssignedBy: dbutil.UUIDPtrToPgUUID(&assignedBy),
	})
	if err != nil {
		return nil, err
	}

	return &RoleAssignmentResponse{
		ID:         res.ID,
		UserID:     res.UserID,
		RoleID:     res.RoleID,
		AssignedAt: res.AssignedAt.Time,
	}, nil
}

func (s *service) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	err := s.repo.RemoveRoleFromUser(ctx, userID, roleID)
	return err
}

func (s *service) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	claims, err := s.jwtManager.ParseRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, ErrInvalidToken
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if user.IsActive == nil || !*user.IsActive {
		return nil, ErrUserInactive
	}

	roles, _ := s.repo.GetUserRoles(ctx, userID)
	roleCodes := make([]string, 0, len(roles))

	for _, r := range roles {
		roleCodes = append(roleCodes, r.Code)
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

	refresh, err := s.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refresh,
		TokenType:    "Bearer",
		ExpiresIn:    900,
	}, nil
}

func (s *service) GetProfile(ctx context.Context, userID uuid.UUID) (*UserProfile, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	roles, _ := s.repo.GetUserRoles(ctx, userID)
	roleInfos := make([]RoleInfo, 0, len(roles))

	for _, r := range roles {
		roleInfos = append(roleInfos, RoleInfo{
			ID:   r.ID,
			Code: r.Code,
			Name: r.Name,
		})
	}

	menus, err := s.repo.GetUserMenus(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get menus failed: %w", err)
	}

	menuInfos := make([]MenuInfo, 0, len(menus))
	for _, m := range menus {
		menuInfos = append(menuInfos, MenuInfo{
			ID:        m.ID,
			ParentID:  uuidFromPg(m.ParentID),
			Code:      m.Code,
			Name:      m.Name,
			Path:      textToPtr(m.Path),
			Icon:      textToPtr(m.Icon),
			CanCreate: m.CanCreate,
			CanRead:   m.CanRead,
			CanUpdate: m.CanUpdate,
			CanDelete: m.CanDelete,
		})
	}

	return &UserProfile{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		FullName:    user.FullName,
		IsActive:    user.IsActive != nil && *user.IsActive,
		LastLoginAt: timeFromPg(user.LastLoginAt),
		CreatedAt:   user.CreatedAt.Time,
		Roles:       roleInfos,
		Menus:       menuInfos,
	}, nil
}

func (s *service) Logout(ctx context.Context, userID uuid.UUID) error {
	// Token blacklist / revoke
	return nil
}

func textToPtr(t *string) *string {
	if t == nil || *t == "" {
		return nil
	}
	return t
}

func timeToPtr(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	return t
}

func timeFromPg(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}
