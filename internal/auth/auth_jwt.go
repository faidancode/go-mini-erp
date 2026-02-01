// internal/auth/auth_jwt.go
package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims is JWT payload used across auth
type Claims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username,omitempty"`
	Email    string   `json:"email,omitempty"`
	Roles    []string `json:"roles,omitempty"`
	jwt.RegisteredClaims
}

// JWTManager defines JWT operations (easy to mock)
type JWTManager interface {
	GenerateAccessToken(userID uuid.UUID, username, email string, roles []string) (string, error)
	GenerateRefreshToken(userID uuid.UUID) (string, error)
	ParseRefreshToken(token string) (*Claims, error)
}

// jwtManager is concrete implementation
type jwtManager struct {
	secret []byte
}

// NewJWTManager creates JWT manager with secret
func NewJWTManager(secret string) JWTManager {
	return &jwtManager{
		secret: []byte(secret),
	}
}

// GenerateAccessToken creates short-lived access token
func (j *jwtManager) GenerateAccessToken(
	userID uuid.UUID,
	username, email string,
	roles []string,
) (string, error) {

	claims := Claims{
		UserID:   userID.String(),
		Username: username,
		Email:    email,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString(j.secret)
}

// GenerateRefreshToken creates long-lived refresh token
func (j *jwtManager) GenerateRefreshToken(userID uuid.UUID) (string, error) {
	claims := Claims{
		UserID: userID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString(j.secret)
}

// ParseRefreshToken validates and parses refresh token
func (j *jwtManager) ParseRefreshToken(token string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return j.secret, nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, ErrInvalidToken
	}

	if claims.ExpiresAt.Before(time.Now()) {
		return nil, ErrTokenExpired
	}

	return claims, nil
}
