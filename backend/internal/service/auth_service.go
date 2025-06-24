package service

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AuthService provides simple username/password authentication and JWT tokens.
// Credentials are defined in configuration.
type AuthService struct {
	username string
	password string
	secret   string
}

func NewAuthService(username, password, secret string) *AuthService {
	return &AuthService{username: username, password: password, secret: secret}
}

// Authenticate verifies credentials and returns a signed JWT on success.
func (a *AuthService) Authenticate(ctx context.Context, user, pass string) (string, error) {
	if user != a.username || pass != a.password {
		return "", fmt.Errorf("invalid credentials")
	}
	claims := jwt.MapClaims{
		"sub": user,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.secret))
}

// Verify parses and validates the token string.
func (a *AuthService) Verify(ctx context.Context, tokenStr string) (*jwt.Token, error) {
	return jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(a.secret), nil
	})
}
