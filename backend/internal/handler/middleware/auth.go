package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/media-parser/backend/internal/model/dto"
	"github.com/media-parser/backend/internal/model/entity"
	"github.com/media-parser/backend/internal/service"
)

type AuthMiddleware struct {
	authService *service.AuthService
}

func NewAuthMiddleware(authService *service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("X-Auth-Token")
		if token == "" {
			c.JSON(http.StatusUnauthorized, dto.Response{
				Success: false,
				Error:   &dto.ErrorData{Code: "UNAUTHORIZED", Message: "Missing X-Auth-Token header"},
			})
			c.Abort()
			return
		}

		t, err := m.authService.ValidateToken(c, token)
		if err != nil {
			status := http.StatusUnauthorized
			errorCode := "UNAUTHORIZED"

			switch err {
			case service.ErrTokenNotFound:
				errorCode = "TOKEN_NOT_FOUND"
			case service.ErrTokenInactive:
				errorCode = "TOKEN_INACTIVE"
			case service.ErrTokenExpired:
				errorCode = "TOKEN_EXPIRED"
			}

			c.JSON(status, dto.Response{
				Success: false,
				Error:   &dto.ErrorData{Code: errorCode, Message: err.Error()},
			})
			c.Abort()
			return
		}

		c.Set("token", t)
		c.Set("token_id", t.ID)

		c.Next()
	}
}

func (m *AuthMiddleware) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenVal, exists := c.Get("token")
		if !exists {
			c.JSON(http.StatusUnauthorized, dto.Response{
				Success: false,
				Error:   &dto.ErrorData{Code: "UNAUTHORIZED", Message: "Token not found in context"},
			})
			c.Abort()
			return
		}

		token, ok := tokenVal.(*entity.APIToken)
		if !ok {
			c.JSON(http.StatusInternalServerError, dto.Response{
				Success: false,
				Error:   &dto.ErrorData{Code: "INTERNAL_ERROR", Message: "Invalid token type"},
			})
			c.Abort()
			return
		}

		if !m.authService.CheckPermission(token, permission) {
			c.JSON(http.StatusForbidden, dto.Response{
				Success: false,
				Error:   &dto.ErrorData{Code: "FORBIDDEN", Message: "Insufficient permissions"},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
