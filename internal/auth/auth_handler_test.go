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

// Test Login - Invalid Credentials
func TestLoginHandler_InvalidCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	handler := auth.NewHandler(mockService)

	router := gin.Default()
	router.POST("/auth/login", handler.Login)

	mockService.EXPECT().
		Login(gomock.Any(), auth.LoginRequest{
			Email:    "test@example.com",
			Password: "wrongpass",
		}).
		Return(nil, auth.ErrInvalidCredentials).
		Times(1)

	// Create request
	body := map[string]string{
		"username": "wronguser",
		"password": "wrongpass",
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "invalid")
}

// Test Login - Bad Request (Missing Fields)
func TestLoginHandler_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	handler := auth.NewHandler(mockService)

	router := gin.Default()
	router.POST("/auth/login", handler.Login)

	// Create request with missing password
	body := map[string]string{
		"username": "testuser",
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test Register - Success
func TestRegisterHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	handler := auth.NewHandler(mockService)

	router := gin.Default()
	router.POST("/auth/register", handler.Register)

	userID := uuid.New()
	expectedResponse := &auth.RegisterResponse{
		ID:        userID,
		Username:  "newuser",
		Email:     "new@example.com",
		FullName:  "New User",
		CreatedAt: time.Now(),
	}

	mockService.EXPECT().
		Register(gomock.Any(), auth.RegisterRequest{
			Username: "newuser",
			Email:    "new@example.com",
			Password: "password123",
			FullName: "New User",
		}).
		Return(expectedResponse, nil).
		Times(1)

	// Create request
	body := map[string]string{
		"username":  "newuser",
		"email":     "new@example.com",
		"password":  "password123",
		"full_name": "New User",
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)

	var response auth.RegisterResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "newuser", response.Username)
}

// Test Register - Username Exists
func TestRegisterHandler_UsernameExists(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	handler := auth.NewHandler(mockService)

	router := gin.Default()
	router.POST("/auth/register", handler.Register)

	mockService.EXPECT().
		Register(gomock.Any(), gomock.Any()).
		Return(nil, auth.ErrUsernameExists).
		Times(1)

	// Create request
	body := map[string]string{
		"username":  "existinguser",
		"email":     "new@example.com",
		"password":  "password123",
		"full_name": "New User",
	}
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusConflict, w.Code)
}

// Test RefreshToken - Success
func TestRefreshTokenHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	handler := auth.NewHandler(mockService)

	router := gin.Default()
	router.POST("/auth/refresh", handler.RefreshToken)

	expectedResponse := &auth.TokenResponse{
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
		TokenType:    "Bearer",
		ExpiresIn:    900,
	}

	mockService.EXPECT().
		RefreshToken(gomock.Any(), "old-refresh-token").
		Return(expectedResponse, nil).
		Times(1)

	// Create request with refresh token in cookie
	req, _ := http.NewRequest("POST", "/auth/refresh", nil)
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: "old-refresh-token",
	})

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response auth.TokenResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "new-access-token", response.AccessToken)
}

// Test RefreshToken - Missing Cookie
func TestRefreshTokenHandler_MissingCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	handler := auth.NewHandler(mockService)

	router := gin.Default()
	router.POST("/auth/refresh", handler.RefreshToken)

	// Create request without cookie
	req, _ := http.NewRequest("POST", "/auth/refresh", nil)

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// Test RefreshToken - Invalid Token
func TestRefreshTokenHandler_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	handler := auth.NewHandler(mockService)

	router := gin.Default()
	router.POST("/auth/refresh", handler.RefreshToken)

	mockService.EXPECT().
		RefreshToken(gomock.Any(), "invalid-token").
		Return(nil, auth.ErrInvalidToken).
		Times(1)

	// Create request
	req, _ := http.NewRequest("POST", "/auth/refresh", nil)
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: "invalid-token",
	})

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// Test GetProfile - Success
func TestGetProfileHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	handler := auth.NewHandler(mockService)

	router := gin.Default()

	// Mock auth middleware - set user_id in context
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "f0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11")
		c.Next()
	})

	router.GET("/auth/profile", handler.GetProfile)

	userID := uuid.MustParse("f0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11")
	expectedResponse := &auth.UserProfile{
		ID:        userID,
		Username:  "testuser",
		Email:     "test@example.com",
		FullName:  "Test User",
		IsActive:  true,
		CreatedAt: time.Now(),
		Roles: []auth.RoleInfo{
			{
				ID:   uuid.New(),
				Code: "admin",
				Name: "Administrator",
			},
		},
		Menus: []auth.MenuInfo{
			{
				ID:        uuid.New(),
				Code:      "dashboard",
				Name:      "Dashboard",
				CanRead:   true,
				CanCreate: true,
				CanUpdate: true,
				CanDelete: true,
			},
		},
	}

	mockService.EXPECT().
		GetProfile(gomock.Any(), userID).
		Return(expectedResponse, nil).
		Times(1)

	// Create request
	req, _ := http.NewRequest("GET", "/auth/profile", nil)

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response auth.UserProfile
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "testuser", response.Username)
	assert.Len(t, response.Roles, 1)
	assert.Len(t, response.Menus, 1)
}

// Test GetProfile - Invalid User ID
func TestGetProfileHandler_InvalidUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	handler := auth.NewHandler(mockService)

	router := gin.Default()

	// Mock auth middleware - set invalid user_id
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "invalid-uuid")
		c.Next()
	})

	router.GET("/auth/profile", handler.GetProfile)

	// Create request
	req, _ := http.NewRequest("GET", "/auth/profile", nil)

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test Logout - Success
func TestLogoutHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	handler := auth.NewHandler(mockService)

	router := gin.Default()

	// Mock auth middleware
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "f0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11")
		c.Next()
	})

	router.POST("/auth/logout", handler.Logout)

	userID := uuid.MustParse("f0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11")

	mockService.EXPECT().
		Logout(gomock.Any(), userID).
		Return(nil).
		Times(1)

	// Create request
	req, _ := http.NewRequest("POST", "/auth/logout", nil)

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Logged out successfully", response["message"])

	// Check if cookie is cleared
	cookies := w.Result().Cookies()
	found := false
	for _, cookie := range cookies {
		if cookie.Name == "refresh_token" {
			found = true
			assert.Equal(t, -1, cookie.MaxAge)
		}
	}
	assert.True(t, found, "refresh_token cookie should be set to expire")
}
