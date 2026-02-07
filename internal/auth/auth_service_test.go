// internal/auth/auth_service_test.go
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

func TestLogin_InvalidUsername(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jwtStub := &jwtManagerStub{}
	repo := mocks.NewMockRepository(ctrl)
	service := auth.NewService(repo, nil, jwtStub)

	repo.EXPECT().
		GetUserByUsername(gomock.Any(), "wronguser").
		Return(db.GetUserByUsernameRow{}, pgx.ErrNoRows)

	result, err := service.Login(context.Background(), auth.LoginRequest{
		Email:    "test@example.com",
		Password: "password",
	})

	assert.ErrorIs(t, err, auth.ErrInvalidCredentials)
	assert.Nil(t, result)
}

func TestLogin_WrongPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jwtStub := &jwtManagerStub{}
	repo := mocks.NewMockRepository(ctrl)
	service := auth.NewService(repo, nil, jwtStub)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)

	repo.EXPECT().
		GetUserByUsername(gomock.Any(), "testuser").
		Return(db.GetUserByUsernameRow{
			ID:           uuid.New(),
			Username:     "testuser",
			PasswordHash: string(hashed),
			IsActive:     dbutil.BoolPtr(true),
		}, nil)

	result, err := service.Login(context.Background(), auth.LoginRequest{
		Email:    "test@example.com",
		Password: "wrong",
	})

	assert.ErrorIs(t, err, auth.ErrInvalidCredentials)
	assert.Nil(t, result)
}

func TestLogin_UserInactive(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jwtStub := &jwtManagerStub{}
	repo := mocks.NewMockRepository(ctrl)
	service := auth.NewService(repo, nil, jwtStub)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)

	repo.EXPECT().
		GetUserByUsername(gomock.Any(), "testuser").
		Return(db.GetUserByUsernameRow{
			ID:           uuid.New(),
			Username:     "testuser",
			PasswordHash: string(hashed),
			IsActive:     dbutil.BoolPtr(false),
		}, nil)

	result, err := service.Login(context.Background(), auth.LoginRequest{
		Email:    "test@example.com",
		Password: "password",
	})

	assert.ErrorIs(t, err, auth.ErrUserInactive)
	assert.Nil(t, result)
}

// =======================
// REGISTER
// =======================

func TestRegister_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jwtStub := &jwtManagerStub{}
	repo := mocks.NewMockRepository(ctrl)
	service := auth.NewService(repo, nil, jwtStub)

	userID := uuid.New()

	repo.EXPECT().
		CheckUsernameExists(gomock.Any(), "newuser").
		Return(false, nil)

	repo.EXPECT().
		CheckEmailExists(gomock.Any(), "new@example.com").
		Return(false, nil)

	repo.EXPECT().
		CreateUser(gomock.Any(), gomock.Any()).
		Return(db.CreateUserRow{
			ID:        userID,
			Username:  "newuser",
			Email:     "new@example.com",
			FullName:  "New User",
			IsActive:  dbutil.BoolPtr(true),
			CreatedAt: dbutil.TimeToPgTime(time.Now()),
		}, nil)

	result, err := service.Register(context.Background(), auth.RegisterRequest{
		Username: "newuser",
		Email:    "new@example.com",
		Password: "password123",
		FullName: "New User",
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "newuser", result.Username)
	assert.Equal(t, "new@example.com", result.Email)
}

func TestRegister_UsernameExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepository(ctrl)
	service := auth.NewService(repo, nil, &jwtManagerStub{})

	repo.EXPECT().
		CheckUsernameExists(gomock.Any(), "existing").
		Return(true, nil)

	result, err := service.Register(context.Background(), auth.RegisterRequest{
		Username: "existing",
		Email:    "x@y.com",
		Password: "password",
	})

	assert.ErrorIs(t, err, auth.ErrUsernameExists)
	assert.Nil(t, result)
}

// =======================
// PROFILE
// =======================

func TestGetProfile_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepository(ctrl)
	service := auth.NewService(repo, nil, &jwtManagerStub{})

	userID := uuid.New()

	repo.EXPECT().
		GetUserByID(gomock.Any(), userID).
		Return(db.GetUserByIDRow{
			ID:        userID,
			Username:  "testuser",
			Email:     "test@example.com",
			FullName:  "Test User",
			IsActive:  dbutil.BoolPtr(true),
			CreatedAt: dbutil.TimeToPgTime(time.Now()),
		}, nil)

	repo.EXPECT().
		GetUserRoles(gomock.Any(), userID).
		Return([]db.GetUserRolesRow{
			{ID: uuid.New(), Code: "admin", Name: "Administrator"},
		}, nil)

	repo.EXPECT().
		GetUserMenus(gomock.Any(), userID).
		Return([]db.GetUserMenusRow{
			{
				ID:        uuid.New(),
				Code:      "dashboard",
				Name:      "Dashboard",
				CanRead:   *dbutil.BoolPtr(true),
				CanCreate: *dbutil.BoolPtr(true),
				CanUpdate: *dbutil.BoolPtr(true),
				CanDelete: *dbutil.BoolPtr(true),
			},
		}, nil)

	result, err := service.GetProfile(context.Background(), userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "testuser", result.Username)
	assert.Len(t, result.Roles, 1)
	assert.Len(t, result.Menus, 1)
	assert.True(t, result.Menus[0].CanRead)
}

func TestGetProfile_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepository(ctrl)
	service := auth.NewService(repo, nil, &jwtManagerStub{})

	repo.EXPECT().
		GetUserByID(gomock.Any(), gomock.Any()).
		Return(db.GetUserByIDRow{}, pgx.ErrNoRows)

	result, err := service.GetProfile(context.Background(), uuid.New())

	assert.ErrorIs(t, err, auth.ErrUserNotFound)
	assert.Nil(t, result)
}

func TestAssignRoleToUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepository(ctrl)
	service := auth.NewService(repo, nil, nil)

	ctx := context.Background()
	userID := uuid.New()
	roleID := uuid.New()
	assignedBy := uuid.New()

	// mock get user by id
	repo.EXPECT().GetUserByID(ctx, userID).Return(db.GetUserByIDRow{
		ID:       userID,
		Username: "user1",
		Email:    "user1@example.com",
		FullName: "User One",
		IsActive: dbutil.BoolPtr(true),
	}, nil)

	// mock assign role
	repo.EXPECT().AssignRoleToUser(ctx, db.AssignRoleToUserParams{
		UserID:     userID,
		RoleID:     roleID,
		AssignedBy: dbutil.UUIDPtrToPgUUID(&assignedBy),
	}).Return(db.AssignRoleToUserRow{
		ID:         uuid.New(),
		UserID:     userID,
		RoleID:     roleID,
		AssignedAt: dbutil.TimeToPgTime(time.Now()),
	}, nil)

	resp, err := service.AssignRoleToUser(ctx, userID, roleID, assignedBy)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, userID, resp.UserID)
	assert.Equal(t, roleID, resp.RoleID)
}

func TestAssignRoleToUser_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepository(ctrl)
	service := auth.NewService(repo, nil, nil)

	ctx := context.Background()
	userID := uuid.New()
	roleID := uuid.New()
	assignedBy := uuid.New()

	repo.EXPECT().GetUserByID(ctx, userID).Return(db.GetUserByIDRow{}, auth.ErrUserNotFound)

	resp, err := service.AssignRoleToUser(ctx, userID, roleID, assignedBy)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, auth.ErrUserNotFound)
}
