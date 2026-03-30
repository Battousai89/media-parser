package service

import (
	"context"

	"github.com/media-parser/backend/internal/model/entity"
	"github.com/media-parser/backend/internal/repository"
)

type DictionaryService struct {
	dictRepo repository.DictionaryRepository
}

func NewDictionaryService(dictRepo repository.DictionaryRepository) *DictionaryService {
	return &DictionaryService{
		dictRepo: dictRepo,
	}
}

func (s *DictionaryService) GetMediaTypes(ctx context.Context) ([]*entity.MediaType, error) {
	return s.dictRepo.GetMediaTypes(ctx)
}

func (s *DictionaryService) GetRequestStatuses(ctx context.Context) ([]*entity.RequestStatus, error) {
	return s.dictRepo.GetRequestStatuses(ctx)
}

func (s *DictionaryService) GetSourceStatuses(ctx context.Context) ([]*entity.SourceStatus, error) {
	return s.dictRepo.GetSourceStatuses(ctx)
}
