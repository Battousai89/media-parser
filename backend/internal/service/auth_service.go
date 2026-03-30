package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/media-parser/backend/internal/model/entity"
	"github.com/media-parser/backend/internal/repository"
)

var (
	ErrTokenNotFound     = errors.New("token not found")
	ErrTokenInactive     = errors.New("token is inactive")
	ErrTokenExpired      = errors.New("token has expired")
	ErrInsufficientPerms = errors.New("insufficient permissions")
)

type AuthService struct {
	tokenRepo repository.APITokenRepository
}

func NewAuthService(tokenRepo repository.APITokenRepository) *AuthService {
	return &AuthService{
		tokenRepo: tokenRepo,
	}
}

func (s *AuthService) ValidateToken(ctx context.Context, token string) (*entity.APIToken, error) {
	t, err := s.tokenRepo.GetByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, ErrTokenNotFound
	}
	if !t.Active {
		return nil, ErrTokenInactive
	}
	if t.ExpiresAt != nil && time.Now().After(*t.ExpiresAt) {
		return nil, ErrTokenExpired
	}

	_ = s.tokenRepo.UpdateLastUsed(ctx, t.ID)

	return t, nil
}

func (s *AuthService) CheckPermission(token *entity.APIToken, permission string) bool {
	if token.Permissions == nil {
		return true
	}

	var perms entity.TokenPermissions
	if err := json.Unmarshal(token.Permissions, &perms); err != nil {
		return true
	}

	switch permission {
	case "parse":
		return perms.Parse
	case "media_read":
		return perms.MediaRead
	case "requests_view":
		return perms.RequestsView
	default:
		return true
	}
}

func (s *AuthService) CreateToken(ctx context.Context, name string, expiresAt *time.Time, perms *entity.TokenPermissions) (*entity.APIToken, error) {
	token := &entity.APIToken{
		Token:     generateToken(),
		Name:      &name,
		Active:    true,
		ExpiresAt: expiresAt,
	}

	if perms != nil {
		data, err := json.Marshal(perms)
		if err != nil {
			return nil, err
		}
		token.Permissions = data
	}

	if err := s.tokenRepo.Create(ctx, token); err != nil {
		return nil, err
	}

	return token, nil
}

func (s *AuthService) RevokeToken(ctx context.Context, id int) error {
	token, err := s.tokenRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if token == nil {
		return ErrTokenNotFound
	}

	token.Active = false
	return s.tokenRepo.Update(ctx, token)
}

func generateToken() string {
	return uuid.New().String() + uuid.New().String()[:8]
}
