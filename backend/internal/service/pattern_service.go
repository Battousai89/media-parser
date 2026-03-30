package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/media-parser/backend/internal/model/entity"
	"github.com/media-parser/backend/internal/repository"
)

var (
	ErrPatternNotFound     = errors.New("pattern not found")
	ErrInvalidRegexPattern = errors.New("invalid regex pattern")
)

type PatternService struct {
	patternRepo repository.PatternRepository
	dictRepo    repository.DictionaryRepository
}

func NewPatternService(
	patternRepo repository.PatternRepository,
	dictRepo repository.DictionaryRepository,
) *PatternService {
	return &PatternService{
		patternRepo: patternRepo,
		dictRepo:    dictRepo,
	}
}

func (s *PatternService) GetAll(ctx context.Context) ([]*entity.Pattern, error) {
	return s.patternRepo.GetAll(ctx)
}

func (s *PatternService) GetByID(ctx context.Context, id int) (*entity.Pattern, error) {
	pattern, err := s.patternRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if pattern == nil {
		return nil, ErrPatternNotFound
	}
	return pattern, nil
}

func (s *PatternService) GetByMediaType(ctx context.Context, mediaTypeCode string) ([]*entity.Pattern, error) {
	mediaType, err := s.dictRepo.GetMediaTypeByCode(ctx, mediaTypeCode)
	if err != nil {
		return nil, err
	}
	if mediaType == nil {
		return nil, errors.New("invalid media type code")
	}

	return s.patternRepo.GetByMediaType(ctx, mediaType.ID)
}

func (s *PatternService) Create(ctx context.Context, name, regex string, mediaTypeID int, priority *int) (*entity.Pattern, error) {
	if _, err := regexp.Compile(regex); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidRegexPattern, err)
	}

	p := 0
	if priority != nil {
		p = *priority
	}

	pattern := &entity.Pattern{
		Name:        name,
		Regex:       regex,
		MediaTypeID: mediaTypeID,
		Priority:    p,
	}

	if err := s.patternRepo.Create(ctx, pattern); err != nil {
		return nil, fmt.Errorf("create pattern: %w", err)
	}

	return pattern, nil
}

func (s *PatternService) Update(ctx context.Context, id int, name *string, regex *string, mediaTypeID *int, priority *int) (*entity.Pattern, error) {
	pattern, err := s.patternRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if pattern == nil {
		return nil, ErrPatternNotFound
	}

	if name != nil {
		pattern.Name = *name
	}
	if regex != nil {
		if _, err := regexp.Compile(*regex); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidRegexPattern, err)
		}
		pattern.Regex = *regex
	}
	if mediaTypeID != nil {
		pattern.MediaTypeID = *mediaTypeID
	}
	if priority != nil {
		pattern.Priority = *priority
	}

	if err := s.patternRepo.Update(ctx, pattern); err != nil {
		return nil, err
	}

	return pattern, nil
}

func (s *PatternService) Delete(ctx context.Context, id int) error {
	return s.patternRepo.Delete(ctx, id)
}
