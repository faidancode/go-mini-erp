package auth_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"

	"go-mini-erp/internal/auth"
	"go-mini-erp/internal/auth/mocks"
	db "go-mini-erp/internal/dbgen"
)

// Test Login - Success
func TestLogin_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	service := auth.NewService(mockRepo, nil)

	ctx := context.Background()
	userID := uuid.New()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	// Setup expectations
	mockRepo.EXPECT().
		GetUserByUsername(ctx, "testuser").
		Return(db.GetUserByUsernameRow{
			ID:           userID,
			Username:     "testuser",
			Email:        "test@example.com",
			PasswordHash: string(hashedPassword),
			FullName:     "Test User",
			IsActive:     db.NewNullBool(true),
		}, nil).
		Times(1)

	mockRepo.EXPECT().
		GetUserRoles(ctx, userID).
		Return([]db.GetUserRolesRow{
			{
				ID:   uuid.New(),
				Code: "admin",
				Name: "Administrator",
			},
		}, nil).
		Times(1)

	mockRepo.EXPECT().
		UpdateUserLastLogin(ctx, userID).
		Return(nil).
		Times(1)

	// Execute
	req := auth.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}
	result, err := service.Login(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
	assert.Equal(t, "Bearer", result.TokenType)
	assert.Equal(t, 900, result.ExpiresIn)
	assert.Equal(t, "testuser", result.User.Username)
	assert.Len(t, result.User.Roles, 1)
	assert.Equal(t, "admin", result.User.Roles[0].Code)
}

// Test Login - Invalid Credentials (User Not Found)
func TestLogin_InvalidCredentials(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	service := auth.NewService(mockRepo, nil)

	ctx := context.Background()

	mockRepo.EXPECT().
		GetUserByUsername(ctx, "wronguser").
		Return(db.GetUserByUsernameRow{}, sql.ErrNoRows).
		Times(1)

	// Execute
	req := auth.LoginRequest{
		Username: "wronguser",
		Password: "password123",
	}
	result, err := service.Login(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, auth.ErrInvalidCredentials)
	assert.Nil(t, result)
}

// Test Login - Wrong Password
func TestLogin_WrongPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	service := auth.NewService(mockRepo, nil)

	ctx := context.Background()
	userID := uuid.New()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)

	mockRepo.EXPECT().
		GetUserByUsername(ctx, "testuser").
		Return(db.GetUserByUsernameRow{
			ID:           userID,
			Username:     "testuser",
			Email:        "test@example.com",
			PasswordHash: string(hashedPassword),
			FullName:     "Test User",
			IsActive:     db.NewNullBool(true),
		}, nil).
		Times(1)

	// Execute
	req := auth.LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}
	result, err := service.Login(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, auth.ErrInvalidCredentials)
	assert.Nil(t, result)
}

// Test Login - User Inactive
func TestLogin_UserInactive(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	service := auth.NewService(mockRepo, nil)

	ctx := context.Background()
	userID := uuid.New()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	mockRepo.EXPECT().
		GetUserByUsername(ctx, "testuser").
		Return(db.GetUserByUsernameRow{
			ID:           userID,
			Username:     "testuser",
			Email:        "test@example.com",
			PasswordHash: string(hashedPassword),
			FullName:     "Test User",
			IsActive:     db.NewNullBool(false), // Inactive
		}, nil).
		Times(1)

	// Execute
	req := auth.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}
	result, err := service.Login(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, auth.ErrUserInactive)
	assert.Nil(t, result)
}

// Test Register - Success
func TestRegister_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	service := auth.NewService(mockRepo, nil)

	ctx := context.Background()
	userID := uuid.New()

	mockRepo.EXPECT().
		CheckUsernameExists(ctx, "newuser").
		Return(false, nil).
		Times(1)

	mockRepo.EXPECT().
		CheckEmailExists(ctx, "new@example.com").
		Return(false, nil).
		Times(1)

	mockRepo.EXPECT().
		CreateUser(ctx, gomock.Any()).
		Return(db.CreateUserRow{
			ID:        userID,
			Username:  "newuser",
			Email:     "new@example.com",
			FullName:  "New User",
			IsActive:  db.NewNullBool(true),
			CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		}, nil).
		Times(1)

	// Execute
	req := auth.RegisterRequest{
		Username: "newuser",
		Email:    "new@example.com",
		Password: "password123",
		FullName: "New User",
	}
	result, err := service.Register(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "newuser", result.Username)
	assert.Equal(t, "new@example.com", result.Email)
}

// Test Register - Username Exists
func TestRegister_UsernameExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	service := auth.NewService(mockRepo, nil)

	ctx := context.Background()

	mockRepo.EXPECT().
		CheckUsernameExists(ctx, "existinguser").
		Return(true, nil).
		Times(1)

	// Execute
	req := auth.RegisterRequest{
		Username: "existinguser",
		Email:    "new@example.com",
		Password: "password123",
		FullName: "New User",
	}
	result, err := service.Register(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, auth.ErrUsernameExists)
	assert.Nil(t, result)
}

// Test Register - Email Exists
func TestRegister_EmailExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	service := auth.NewService(mockRepo, nil)

	ctx := context.Background()

	mockRepo.EXPECT().
		CheckUsernameExists(ctx, "newuser").
		Return(false, nil).
		Times(1)

	mockRepo.EXPECT().
		CheckEmailExists(ctx, "existing@example.com").
		Return(true, nil).
		Times(1)

	// Execute
	req := auth.RegisterRequest{
		Username: "newuser",
		Email:    "existing@example.com",
		Password: "password123",
		FullName: "New User",
	}
	result, err := service.Register(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, auth.ErrEmailExists)
	assert.Nil(t, result)
}

// Test GetProfile - Success
func TestGetProfile_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	service := auth.NewService(mockRepo, nil)

	ctx := context.Background()
	userID := uuid.New()

	mockRepo.EXPECT().
		GetUserByID(ctx, userID).
		Return(db.GetUserByIDRow{
			ID:        userID,
			Username:  "testuser",
			Email:     "test@example.com",
			FullName:  "Test User",
			IsActive:  db.NewNullBool(true),
			CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		}, nil).
		Times(1)

	mockRepo.EXPECT().
		GetUserRoles(ctx, userID).
		Return([]db.GetUserRolesRow{
			{
				ID:   uuid.New(),
				Code: "admin",
				Name: "Administrator",
			},
		}, nil).
		Times(1)

	mockRepo.EXPECT().
		GetUserMenus(ctx, userID).
		Return([]db.GetUserMenusRow{
			{
				ID:        uuid.New(),
				Code:      "dashboard",
				Name:      "Dashboard",
				CanCreate: int64(1),
				CanRead:   int64(1),
				CanUpdate: int64(1),
				CanDelete: int64(1),
			},
		}, nil).
		Times(1)

	// Execute
	result, err := service.GetProfile(ctx, userID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "testuser", result.Username)
	assert.Len(t, result.Roles, 1)
	assert.Len(t, result.Menus, 1)
	assert.True(t, result.Menus[0].CanRead)
}

// Test GetProfile - User Not Found
func TestGetProfile_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	service := auth.NewService(mockRepo, nil)

	ctx := context.Background()
	userID := uuid.New()

	mockRepo.EXPECT().
		GetUserByID(ctx, userID).
		Return(db.GetUserByIDRow{}, sql.ErrNoRows).
		Times(1)

	// Execute
	result, err := service.GetProfile(ctx, userID)

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, auth.ErrUserNotFound)
	assert.Nil(t, result)
}
