package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthServiceInterface defines the auth methods used by the handler.
type AuthServiceInterface interface {
	Authenticate(ctx context.Context, user, pass string) (string, error)
	Verify(ctx context.Context, token string) (*jwt.Token, error)
}

// AuthHandler exposes a login endpoint.
type AuthHandler struct {
	svc AuthServiceInterface
}

func NewAuthHandler(s AuthServiceInterface) *AuthHandler {
	return &AuthHandler{svc: s}
}

func (h *AuthHandler) HandleLogin(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, err := h.svc.Authenticate(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func AuthMiddleware(s AuthServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if len(auth) < 7 || auth[:7] != "Bearer " {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		tokenStr := auth[7:]
		if _, err := s.Verify(c.Request.Context(), tokenStr); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Next()
	}
}
