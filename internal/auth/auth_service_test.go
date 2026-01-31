package auth

import (
	"context"
	"database/sql"
	"go-mini-erp/internal/dbgen"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// Mock Repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetUserByUsername(ctx context.Context, username string) (dbgen.GetUserByUsernameRow, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(dbgen.GetUserByUsernameRow), args.Error(1)
}

func (m *MockRepository) GetUserByID(ctx context.Context, id uuid.UUID) (dbgen.GetUserByIDRow, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(dbgen.GetUserByIDRow), args.Error(1)
}

func (m *MockRepository) GetUserByEmail(ctx context.Context, email string) (dbgen.GetUserByEmailRow, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(dbgen.GetUserByEmailRow), args.Error(1)
}

func (m *MockRepository) CreateUser(ctx context.Context, arg dbgen.CreateUserParams) (dbgen.CreateUserRow, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(dbgen.CreateUserRow), args.Error(1)
}

func (m *MockRepository) UpdateUserLastLogin(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]dbgen.GetUserRolesRow, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]dbgen.GetUserRolesRow), args.Error(1)
}

func (m *MockRepository) GetUserMenus(ctx context.Context, userID uuid.UUID) ([]dbgen.GetUserMenusRow, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]dbgen.GetUserMenusRow), args.Error(1)
}

func (m *MockRepository) AssignRoleToUser(ctx context.Context, arg dbgen.AssignRoleToUserParams) (dbgen.AssignRoleToUserRow, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(dbgen.AssignRoleToUserRow), args.Error(1)
}

func (m *MockRepository) CheckUsernameExists(ctx context.Context, username string) (bool, error) {
	args := m.Called(ctx, username)
	return args.Bool(0), args.Error(1)
}

func (m *MockRepository) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

// Test Login - Success
func TestLogin_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := &service{
		repo:      mockRepo,
		jwtSecret: []byte("test-secret"),
	}

	ctx := context.Background()
	userID := uuid.New()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	// Mock GetUserByUsername
	mockRepo.On("GetUserByUsername", ctx, "testuser").Return(dbgen.GetUserByUsernameRow{
		ID:           userID,
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
		FullName:     "Test User",
		IsActive:     dbgen.NewNullBool(true),
	}, nil)

	// Mock GetUserRoles
	mockRepo.On("GetUserRoles", ctx, userID).Return([]dbgen.GetUserRolesRow{
		{
			ID:   uuid.New(),
			Code: "admin",
			Name: "Administrator",
		},
	}, nil)

	// Mock UpdateUserLastLogin
	mockRepo.On("UpdateUserLastLogin", ctx, userID).Return(nil)

	// Execute
	req := LoginRequest{
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

	mockRepo.AssertExpectations(t)
}

// Test Login - Invalid Credentials
func TestLogin_InvalidCredentials(t *testing.T) {
	mockRepo := new(MockRepository)
	service := &service{
		repo:      mockRepo,
		jwtSecret: []byte("test-secret"),
	}

	ctx := context.Background()

	// Mock GetUserByUsername - user not found
	mockRepo.On("GetUserByUsername", ctx, "wronguser").Return(dbgen.GetUserByUsernameRow{}, sql.ErrNoRows)

	// Execute
	req := LoginRequest{
		Username: "wronguser",
		Password: "password123",
	}
	result, err := service.Login(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)
	assert.Nil(t, result)

	mockRepo.AssertExpectations(t)
}

// Test Login - Wrong Password
func TestLogin_WrongPassword(t *testing.T) {
	mockRepo := new(MockRepository)
	service := &service{
		repo:      mockRepo,
		jwtSecret: []byte("test-secret"),
	}

	ctx := context.Background()
	userID := uuid.New()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)

	// Mock GetUserByUsername
	mockRepo.On("GetUserByUsername", ctx, "testuser").Return(dbgen.GetUserByUsernameRow{
		ID:           userID,
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
		FullName:     "Test User",
		IsActive:     dbgen.NewNullBool(true),
	}, nil)

	// Execute
	req := LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}
	result, err := service.Login(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)
	assert.Nil(t, result)

	mockRepo.AssertExpectations(t)
}

// Test Login - User Inactive
func TestLogin_UserInactive(t *testing.T) {
	mockRepo := new(MockRepository)
	service := &service{
		repo:      mockRepo,
		jwtSecret: []byte("test-secret"),
	}

	ctx := context.Background()
	userID := uuid.New()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	// Mock GetUserByUsername - inactive user
	mockRepo.On("GetUserByUsername", ctx, "testuser").Return(dbgen.GetUserByUsernameRow{
		ID:           userID,
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
		FullName:     "Test User",
		IsActive:     dbgen.NewNullBool(false), // Inactive
	}, nil)

	// Execute
	req := LoginRequest{
		Username: "testuser",
		Password: "password123",
	}
	result, err := service.Login(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrUserInactive, err)
	assert.Nil(t, result)

	mockRepo.AssertExpectations(t)
}

// Test Register - Success
func TestRegister_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := &service{
		repo:      mockRepo,
		jwtSecret: []byte("test-secret"),
	}

	ctx := context.Background()
	userID := uuid.New()

	// Mock CheckUsernameExists
	mockRepo.On("CheckUsernameExists", ctx, "newuser").Return(false, nil)

	// Mock CheckEmailExists
	mockRepo.On("CheckEmailExists", ctx, "new@example.com").Return(false, nil)

	// Mock CreateUser
	mockRepo.On("CreateUser", ctx, mock.AnythingOfType("sqlc.CreateUserParams")).Return(dbgen.CreateUserRow{
		ID:        userID,
		Username:  "newuser",
		Email:     "new@example.com",
		FullName:  "New User",
		IsActive:  dbgen.NewNullBool(true),
		CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
	}, nil)

	// Execute
	req := RegisterRequest{
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

	mockRepo.AssertExpectations(t)
}

// Test Register - Username Exists
func TestRegister_UsernameExists(t *testing.T) {
	mockRepo := new(MockRepository)
	service := &service{
		repo:      mockRepo,
		jwtSecret: []byte("test-secret"),
	}

	ctx := context.Background()

	// Mock CheckUsernameExists - username exists
	mockRepo.On("CheckUsernameExists", ctx, "existinguser").Return(true, nil)

	// Execute
	req := RegisterRequest{
		Username: "existinguser",
		Email:    "new@example.com",
		Password: "password123",
		FullName: "New User",
	}
	result, err := service.Register(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrUsernameExists, err)
	assert.Nil(t, result)

	mockRepo.AssertExpectations(t)
}

// Test Register - Email Exists
func TestRegister_EmailExists(t *testing.T) {
	mockRepo := new(MockRepository)
	service := &service{
		repo:      mockRepo,
		jwtSecret: []byte("test-secret"),
	}

	ctx := context.Background()

	// Mock CheckUsernameExists
	mockRepo.On("CheckUsernameExists", ctx, "newuser").Return(false, nil)

	// Mock CheckEmailExists - email exists
	mockRepo.On("CheckEmailExists", ctx, "existing@example.com").Return(true, nil)

	// Execute
	req := RegisterRequest{
		Username: "newuser",
		Email:    "existing@example.com",
		Password: "password123",
		FullName: "New User",
	}
	result, err := service.Register(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrEmailExists, err)
	assert.Nil(t, result)

	mockRepo.AssertExpectations(t)
}

// Test GetProfile - Success
func TestGetProfile_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	service := &service{
		repo:      mockRepo,
		jwtSecret: []byte("test-secret"),
	}

	ctx := context.Background()
	userID := uuid.New()

	// Mock GetUserByID
	mockRepo.On("GetUserByID", ctx, userID).Return(dbgen.GetUserByIDRow{
		ID:        userID,
		Username:  "testuser",
		Email:     "test@example.com",
		FullName:  "Test User",
		IsActive:  dbgen.NewNullBool(true),
		CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
	}, nil)

	// Mock GetUserRoles
	mockRepo.On("GetUserRoles", ctx, userID).Return([]dbgen.GetUserRolesRow{
		{
			ID:   uuid.New(),
			Code: "admin",
			Name: "Administrator",
		},
	}, nil)

	// Mock GetUserMenus
	mockRepo.On("GetUserMenus", ctx, userID).Return([]dbgen.GetUserMenusRow{
		{
			ID:        uuid.New(),
			Code:      "dashboard",
			Name:      "Dashboard",
			CanCreate: int64(1),
			CanRead:   int64(1),
			CanUpdate: int64(1),
			CanDelete: int64(1),
		},
	}, nil)

	// Execute
	result, err := service.GetProfile(ctx, userID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "testuser", result.Username)
	assert.Len(t, result.Roles, 1)
	assert.Len(t, result.Menus, 1)
	assert.True(t, result.Menus[0].CanRead)

	mockRepo.AssertExpectations(t)
}

// Test GetProfile - User Not Found
func TestGetProfile_UserNotFound(t *testing.T) {
	mockRepo := new(MockRepository)
	service := &service{
		repo:      mockRepo,
		jwtSecret: []byte("test-secret"),
	}

	ctx := context.Background()
	userID := uuid.New()

	// Mock GetUserByID - user not found
	mockRepo.On("GetUserByID", ctx, userID).Return(dbgen.GetUserByIDRow{}, sql.ErrNoRows)

	// Execute
	result, err := service.GetProfile(ctx, userID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)
	assert.Nil(t, result)

	mockRepo.AssertExpectations(t)
}
