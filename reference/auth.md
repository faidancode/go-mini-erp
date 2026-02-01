## Service
package auth

import (
	"context"
	"os"
	"time"

	autherrors "gadget-api/internal/auth/errors"
	"gadget-api/internal/db"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Login(ctx context.Context, email, password string) (string, string, AuthResponse, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return "", "", AuthResponse{}, autherrors.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", "", AuthResponse{}, autherrors.ErrInvalidCredentials
	}

	accessToken, err := s.generateToken(user.ID.String(), user.Role, time.Minute*15)
	if err != nil {
		return "", "", AuthResponse{}, autherrors.ErrTokenGenerationFailed
	}

	refreshToken, err := s.generateToken(user.ID.String(), user.Role, time.Hour*24*7)
	if err != nil {
		return "", "", AuthResponse{}, autherrors.ErrTokenGenerationFailed
	}

	return accessToken, refreshToken, AuthResponse{
		ID:    user.ID.String(),
		Email: user.Email,
		Name:  user.Name,
		Role:  user.Role,
	}, nil
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (string, string, AuthResponse, error) {
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, autherrors.ErrInvalidToken
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !token.Valid {
		return "", "", AuthResponse{}, autherrors.ErrInvalidRefreshToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", AuthResponse{}, autherrors.ErrInvalidToken
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return "", "", AuthResponse{}, autherrors.ErrInvalidToken
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return "", "", AuthResponse{}, autherrors.ErrInvalidUserID
	}

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return "", "", AuthResponse{}, autherrors.ErrUserNotFound
	}

	newAccessToken, err := s.generateToken(user.ID.String(), user.Role, time.Minute*15)
	if err != nil {
		return "", "", AuthResponse{}, autherrors.ErrTokenGenerationFailed
	}

	newRefreshToken, err := s.generateToken(user.ID.String(), user.Role, time.Hour*24*7)
	if err != nil {
		return "", "", AuthResponse{}, autherrors.ErrTokenGenerationFailed
	}

	return newAccessToken, newRefreshToken, AuthResponse{
		ID:    user.ID.String(),
		Email: user.Email,
		Name:  user.Name,
		Role:  user.Role,
	}, nil
}

func (s *Service) GetMe(ctx context.Context, userID string) (*AuthResponse, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, autherrors.ErrInvalidUserID
	}

	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, autherrors.ErrUserNotFound
	}

	return &AuthResponse{
		ID:    u.ID.String(),
		Email: u.Email,
		Name:  u.Name,
		Role:  u.Role,
	}, nil
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (AuthResponse, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return AuthResponse{}, autherrors.ErrTokenGenerationFailed
	}

	user, err := s.repo.Create(ctx, db.CreateUserParams{
		Email:    req.Email,
		Name:     req.Name,
		Password: string(hashed),
		Role:     "CUSTOMER",
	})
	if err != nil {
		return AuthResponse{}, autherrors.ErrEmailAlreadyRegistered
	}

	return AuthResponse{
		ID:    user.ID.String(),
		Email: user.Email,
		Name:  user.Name,
		Role:  user.Role,
	}, nil
}

// reusable token generator
func (s *Service) generateToken(userID, role string, expiry time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(expiry).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

## Service Test
package auth_test

import (
	"context"
	"errors"
	"gadget-api/internal/auth"
	"gadget-api/internal/db"
	authMock "gadget-api/internal/mock/auth"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestService_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := authMock.NewMockRepository(ctrl)
	service := auth.NewService(mockRepo)
	ctx := context.Background()

	pw, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	t.Run("Success Login", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByEmail(ctx, "admin").
			Return(db.User{Email: "admin", Password: string(pw)}, nil)

		token, refreshToken, resp, err := service.Login(ctx, "admin", "password123")

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.NotEmpty(t, refreshToken)
		assert.Equal(t, "admin", resp.Email)
	})

	t.Run("Wrong Password", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByEmail(ctx, "admin").
			Return(db.User{Email: "admin", Password: string(pw)}, nil)

		_, _, _, err := service.Login(ctx, "admin", "wrongpass")
		assert.Error(t, err)
	})
}

func TestService_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := authMock.NewMockRepository(ctrl)
	service := auth.NewService(mockRepo)
	ctx := context.Background()

	t.Run("Success Register", func(t *testing.T) {
		req := auth.RegisterRequest{
			Email:    "user@example.com",
			Password: "password123",
		}

		mockRepo.EXPECT().
			Create(ctx, gomock.Any()).
			Return(db.CreateUserRow{
				Email: req.Email,
				Role:  "CUSTOMER",
			}, nil)

		resp, err := service.Register(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, req.Email, resp.Email)
		assert.Equal(t, "CUSTOMER", resp.Role)
	})

	t.Run("Error Register", func(t *testing.T) {
		req := auth.RegisterRequest{
			Email:    "user@example.com",
			Password: "password123",
		}

		mockRepo.EXPECT().
			Create(ctx, gomock.Any()).
			Return(db.CreateUserRow{}, errors.New("duplicate email"))

		_, err := service.Register(ctx, req)
		assert.Error(t, err)
	})
}

## Handler/Controller
package auth

import (
	platform "gadget-api/internal/pkg/request"
	"gadget-api/internal/pkg/response"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	service *Service
}

func NewController(s *Service) *Controller {
	return &Controller{service: s}
}

func (ctrl *Controller) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Response Error Seragam
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Input tidak valid", err.Error())
		return
	}

	clientHeader := c.GetHeader("X-Client-Type")
	userAgent := c.GetHeader("User-Agent")
	clientType := platform.ResolveClientType(clientHeader, userAgent)

	token, refreshToken, userResp, err := ctrl.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		// Response Error Seragam
		response.Error(c, http.StatusUnauthorized, "AUTH_FAILED", "Email atau password salah", nil)
		return
	}
	isProd := os.Getenv("APP_ENV") == "production"

	if platform.IsWebClient(clientType) {
		// Set access_token cookie
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "access_token",
			Value:    token,
			Path:     "/",
			MaxAge:   86400, // 1 hari
			HttpOnly: true,
			Secure:   isProd,
			SameSite: http.SameSiteLaxMode, // ✅ Explicit SameSite
		})

		// Set refresh_token cookie
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "refresh_token",
			Value:    refreshToken,
			Path:     "/",
			MaxAge:   3600 * 24 * 7, // 7 hari
			HttpOnly: true,
			Secure:   isProd,
			SameSite: http.SameSiteLaxMode, // ✅ Explicit SameSite
		})
	}

	responseData := gin.H{
		"user":          userResp,
		"access_token":  token,
		"refresh_token": refreshToken,
	}

	response.Success(c, http.StatusOK, responseData, nil)
}

func (ctrl *Controller) Me(c *gin.Context) {
	// asumsi middleware sudah set userID di context
	log.Printf("auth context: %+v\n", c.Keys)

	userID, ok := c.Get("user_id")
	if !ok {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized", nil)
		return
	}

	userResp, err := ctrl.service.GetMe(
		c.Request.Context(),
		userID.(string),
	)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized", nil)
		return
	}

	response.Success(c, http.StatusOK, userResp, nil)
}

// auth/auth_controller.go

func (ctrl *Controller) Logout(c *gin.Context) {
	// Ambil isProd dari config
	isProd := os.Getenv("APP_ENV") == "production" // atau dari config Anda

	// Clear access_token
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isProd,
		SameSite: http.SameSiteLaxMode, // ✅ Harus sama dengan login
	})

	// Clear refresh_token
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isProd,
		SameSite: http.SameSiteLaxMode, // ✅ Harus sama dengan login
	})

	response.Success(c, http.StatusOK, "Logout success.", nil)
}

func (ctrl *Controller) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Input tidak valid", err.Error())
		return
	}

	res, err := ctrl.service.Register(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "REGISTER_FAILED", err.Error(), nil)
		return
	}

	response.Success(c, http.StatusCreated, res, nil)
}

func (ctrl *Controller) RefreshToken(c *gin.Context) {
	// 1. Deteksi Client
	clientHeader := c.GetHeader("X-Client-Type")
	userAgent := c.GetHeader("User-Agent")
	clientType := platform.ResolveClientType(clientHeader, userAgent)

	var refreshToken string
	isWeb := platform.IsWebClient(clientType)

	// 2. Ambil Refresh Token (Cookie vs Body)
	if isWeb {
		var err error
		refreshToken, err = c.Cookie("refresh_token")
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "NO_REFRESH_TOKEN", "Missing refresh token", nil)
			return
		}
	} else {
		var req struct {
			RefreshToken string `json:"refresh_token" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Refresh token is required", nil)
			return
		}
		refreshToken = req.RefreshToken
	}

	// 3. Panggil Service untuk Verify & Issue New Tokens
	// Mengembalikan accessToken, newRefreshToken, userDetail, error
	newAccess, newRefresh, userResp, err := ctrl.service.RefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "INVALID_TOKEN", err.Error(), nil)
		return
	}

	isProd := os.Getenv("APP_ENV") == "production"

	// 4. Sinkronisasi Web (Set-Cookie)
	if isWeb {
		// Update Access Token di Cookie
		c.SetCookie("access_token", newAccess, 15*60, "/", "", isProd, true)
		// Update Refresh Token di Cookie
		c.SetCookie("refresh_token", newRefresh, 3600*24*7, "/", "", isProd, true)
	}

	// 5. Response Success (Tetap kirim body untuk sinkronisasi state di frontend)
	responseData := gin.H{
		"user":          userResp,
		"access_token":  newAccess,
		"refresh_token": newRefresh,
	}

	response.Success(c, http.StatusOK, responseData, nil)
}
