package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/media-parser/backend/internal/model/dto"
	"github.com/media-parser/backend/internal/model/entity"
)

type URLCacheEntry struct {
	Hash      string    `json:"hash"`
	ParsedAt  time.Time `json:"parsed_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type SourceRepository interface {
	Create(ctx context.Context, source *entity.Source) error
	GetByID(ctx context.Context, id int) (*entity.Source, error)
	GetAll(ctx context.Context) ([]*entity.Source, error)
	GetActive(ctx context.Context) ([]*entity.Source, error)
	Update(ctx context.Context, source *entity.Source) error
	UpdateStatus(ctx context.Context, id int, statusID int) error
	Delete(ctx context.Context, id int) error
	GetByURL(ctx context.Context, url string) (*entity.Source, error)
}

type PatternRepository interface {
	Create(ctx context.Context, pattern *entity.Pattern) error
	GetByID(ctx context.Context, id int) (*entity.Pattern, error)
	GetAll(ctx context.Context) ([]*entity.Pattern, error)
	GetByMediaType(ctx context.Context, mediaTypeID int) ([]*entity.Pattern, error)
	Update(ctx context.Context, pattern *entity.Pattern) error
	Delete(ctx context.Context, id int) error
}

type RequestRepository interface {
	Create(ctx context.Context, req *entity.Request) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Request, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entity.Request, error)
	GetByStatus(ctx context.Context, statusID int, limit, offset int) ([]*entity.Request, error)
	Update(ctx context.Context, req *entity.Request) error
	UpdateStatus(ctx context.Context, id uuid.UUID, statusID int, errorMsg *string) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int, error)
	GetRequestMedia(ctx context.Context, requestID uuid.UUID, limit, offset int) ([]*entity.Media, error)
	GetRequestMediaWithSources(ctx context.Context, requestID uuid.UUID, limit, offset int) ([]*dto.RequestMediaItem, error)
	IncrementRetryCount(ctx context.Context, id uuid.UUID) error
	GetMediaTypeIDs(ctx context.Context, requestID uuid.UUID) ([]int, error)
	SetMediaTypeIDs(ctx context.Context, requestID uuid.UUID, mediaTypeIDs []int) error
}

type RequestSourceRepository interface {
	Create(ctx context.Context, rs *entity.RequestSource) error
	GetByRequestID(ctx context.Context, requestID uuid.UUID) ([]*entity.RequestSource, error)
	GetByRequestIDAndSourceID(ctx context.Context, requestID uuid.UUID, sourceID int) (*entity.RequestSource, error)
	Update(ctx context.Context, rs *entity.RequestSource) error
	UpdateStatus(ctx context.Context, requestID uuid.UUID, sourceID int, statusID int, parsedCount int, errorMsg *string) error
	UpdateParsedCount(ctx context.Context, requestID uuid.UUID, sourceID int, parsedCount int) error
	IncrementRetryCount(ctx context.Context, requestID uuid.UUID, sourceID int) error
}

type MediaRepository interface {
	Create(ctx context.Context, media *entity.Media) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Media, error)
	GetByURL(ctx context.Context, url string) (*entity.Media, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entity.Media, error)
	GetByMediaType(ctx context.Context, mediaTypeID int, limit, offset int) ([]*entity.Media, error)
	GetByHash(ctx context.Context, hash string) (*entity.Media, error)
	Update(ctx context.Context, media *entity.Media) error
	UpdateStoragePath(ctx context.Context, id uuid.UUID, storagePath string) error
	UpdateAvailability(ctx context.Context, id uuid.UUID, available bool) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int, error)
	CountByMediaType(ctx context.Context, mediaTypeID int) (int, error)
}

type SourceMediaRepository interface {
	Create(ctx context.Context, sm *entity.SourceMedia) error
	GetBySourceID(ctx context.Context, sourceID int, limit, offset int) ([]*entity.SourceMedia, error)
	GetByRequestID(ctx context.Context, requestID uuid.UUID, limit, offset int) ([]*entity.SourceMedia, error)
	GetByMediaID(ctx context.Context, mediaID uuid.UUID) ([]*entity.SourceMedia, error)
}

type DictionaryRepository interface {
	GetRequestStatuses(ctx context.Context) ([]*entity.RequestStatus, error)
	GetMediaTypes(ctx context.Context) ([]*entity.MediaType, error)
	GetSourceStatuses(ctx context.Context) ([]*entity.SourceStatus, error)
	GetRequestStatusByCode(ctx context.Context, code string) (*entity.RequestStatus, error)
	GetMediaTypeByCode(ctx context.Context, code string) (*entity.MediaType, error)
	GetSourceStatusByCode(ctx context.Context, code string) (*entity.SourceStatus, error)
}

type MediaTypeExtensionRepository interface {
	GetByMediaTypeID(ctx context.Context, mediaTypeID int) ([]*entity.MediaTypeExtension, error)
	GetMediaTypeByExtension(ctx context.Context, extension string) (*int, error)
}

type APITokenRepository interface {
	Create(ctx context.Context, token *entity.APIToken) error
	GetByToken(ctx context.Context, token string) (*entity.APIToken, error)
	GetByID(ctx context.Context, id int) (*entity.APIToken, error)
	UpdateLastUsed(ctx context.Context, id int) error
	Update(ctx context.Context, token *entity.APIToken) error
	Delete(ctx context.Context, id int) error
}

type URLCacheRepository interface {
	GetByURL(ctx context.Context, url string) (*entity.URLCache, error)
	DeleteExpired(ctx context.Context) error
	Count(ctx context.Context) (int, error)
}

type RedisCacheRepository interface {
	IsURLCached(ctx context.Context, url string) (bool, error)
	GetURLCache(ctx context.Context, url string) (*URLCacheEntry, error)
	SetURLCache(ctx context.Context, url, hash string, ttl time.Duration) error
	DeleteURLCache(ctx context.Context, url string) error
	Close() error
	Ping(ctx context.Context) error
}
