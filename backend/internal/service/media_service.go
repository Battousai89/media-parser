package service

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/media-parser/backend/internal/model/entity"
	"github.com/media-parser/backend/internal/repository"
)

var (
	ErrMediaNotFound = errors.New("media not found")
)

type MediaService struct {
	mediaRepo       repository.MediaRepository
	sourceMediaRepo repository.SourceMediaRepository
	dictRepo        repository.DictionaryRepository
}

func NewMediaService(
	mediaRepo repository.MediaRepository,
	sourceMediaRepo repository.SourceMediaRepository,
	dictRepo repository.DictionaryRepository,
) *MediaService {
	return &MediaService{
		mediaRepo:       mediaRepo,
		sourceMediaRepo: sourceMediaRepo,
		dictRepo:        dictRepo,
	}
}

func (s *MediaService) GetMediaByID(ctx context.Context, id uuid.UUID) (*entity.Media, error) {
	return s.mediaRepo.GetByID(ctx, id)
}

func (s *MediaService) GetAllMedia(ctx context.Context, limit, offset int) ([]*entity.Media, error) {
	return s.mediaRepo.GetAll(ctx, limit, offset)
}

func (s *MediaService) GetMediaByType(ctx context.Context, mediaTypeCode string, limit, offset int) ([]*entity.Media, error) {
	mediaType, err := s.dictRepo.GetMediaTypeByCode(ctx, mediaTypeCode)
	if err != nil {
		return nil, err
	}
	if mediaType == nil {
		return nil, errors.New("invalid media type code")
	}

	return s.mediaRepo.GetByMediaType(ctx, mediaType.ID, limit, offset)
}

func (s *MediaService) CreateMedia(
	ctx context.Context,
	url string,
	mediaTypeCode string,
	title *string,
	description *string,
	fileSize *int64,
	mimeType *string,
	hash *string,
	meta map[string]interface{},
) (*entity.Media, error) {
	mediaType, err := s.dictRepo.GetMediaTypeByCode(ctx, mediaTypeCode)
	if err != nil {
		return nil, err
	}
	if mediaType == nil {
		return nil, errors.New("invalid media type code")
	}

	media := &entity.Media{
		ID:          uuid.New(),
		URL:         url,
		MediaTypeID: mediaType.ID,
		MediaType:   mediaType,
		Title:       title,
		Description: description,
		FileSize:    fileSize,
		MimeType:    mimeType,
		Hash:        hash,
		Meta:        nil,
		Available:   true,
	}

	if meta != nil {
		metaJSON, _ := json.Marshal(meta)
		media.Meta = metaJSON
	}

	if err := s.mediaRepo.Create(ctx, media); err != nil {
		return nil, err
	}

	return media, nil
}

func (s *MediaService) UpdateMediaAvailability(ctx context.Context, id uuid.UUID, available bool) error {
	return s.mediaRepo.UpdateAvailability(ctx, id, available)
}

func (s *MediaService) DeleteMedia(ctx context.Context, id uuid.UUID) error {
	return s.mediaRepo.Delete(ctx, id)
}

func (s *MediaService) LinkMediaToSource(
	ctx context.Context,
	sourceID int,
	mediaID uuid.UUID,
	requestID *uuid.UUID,
) (*entity.SourceMedia, error) {
	sm := &entity.SourceMedia{
		SourceID:  sourceID,
		MediaID:   mediaID,
		RequestID: requestID,
	}

	if err := s.sourceMediaRepo.Create(ctx, sm); err != nil {
		return nil, err
	}

	return sm, nil
}

func (s *MediaService) GetMediaBySource(ctx context.Context, sourceID int, limit, offset int) ([]*entity.SourceMedia, error) {
	return s.sourceMediaRepo.GetBySourceID(ctx, sourceID, limit, offset)
}

func (s *MediaService) GetMediaByRequest(ctx context.Context, requestID uuid.UUID, limit, offset int) ([]*entity.SourceMedia, error) {
	return s.sourceMediaRepo.GetByRequestID(ctx, requestID, limit, offset)
}

func (s *MediaService) CountMedia(ctx context.Context) (int, error) {
	return s.mediaRepo.Count(ctx)
}

func (s *MediaService) CountMediaByType(ctx context.Context, mediaTypeCode string) (int, error) {
	mediaType, err := s.dictRepo.GetMediaTypeByCode(ctx, mediaTypeCode)
	if err != nil {
		return 0, err
	}
	if mediaType == nil {
		return 0, errors.New("invalid media type code")
	}

	return s.mediaRepo.CountByMediaType(ctx, mediaType.ID)
}
