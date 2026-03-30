package service

import (
	"context"
	"errors"
	"log"

	"github.com/media-parser/backend/internal/model/entity"
	"github.com/media-parser/backend/internal/queue"
	"github.com/media-parser/backend/internal/repository"
)

var (
	ErrParseFailed = errors.New("parse failed")
)

type ParseService struct {
	requestService  *RequestService
	mediaService    *MediaService
	sourceService   *SourceService
	sourceRepo      repository.SourceRepository
	patternRepo     repository.PatternRepository
	mediaRepo       repository.MediaRepository
	sourceMediaRepo repository.SourceMediaRepository
	cacheRepo       repository.RedisCacheRepository
	producer        *queue.Producer
}

func NewParseService(
	requestService *RequestService,
	mediaService *MediaService,
	sourceService *SourceService,
	sourceRepo repository.SourceRepository,
	patternRepo repository.PatternRepository,
	mediaRepo repository.MediaRepository,
	sourceMediaRepo repository.SourceMediaRepository,
	cacheRepo repository.RedisCacheRepository,
	producer *queue.Producer,
) *ParseService {
	log.Printf("ParseService: producer connected = %v", producer != nil)
	return &ParseService{
		requestService:  requestService,
		mediaService:    mediaService,
		sourceService:   sourceService,
		sourceRepo:      sourceRepo,
		patternRepo:     patternRepo,
		mediaRepo:       mediaRepo,
		sourceMediaRepo: sourceMediaRepo,
		cacheRepo:       cacheRepo,
		producer:        producer,
	}
}

func (s *ParseService) ParseURL(
	ctx context.Context,
	url string,
	mediaTypeCodes []string,
	limit, offset, priority int,
) (*entity.Request, error) {
	req, err := s.requestService.CreateRequest(ctx, mediaTypeCodes, limit, offset, priority, nil)
	if err != nil {
		return nil, err
	}

	source, err := s.sourceService.GetOrCreateSource(ctx, url)
	if err != nil {
		return nil, err
	}

	_, err = s.requestService.AddSourceToRequest(ctx, req.ID, source.ID, limit)
	if err != nil {
		return nil, err
	}

	task := &queue.ParseTask{
		RequestID:  req.ID,
		SourceID:   source.ID,
		URL:        url,
		MediaType:  nil,
		Limit:      limit,
		Priority:   priority,
		RetryCount: 0,
		MaxRetries: 3,
		CreatedAt:  req.CreatedAt,
	}

	if s.producer != nil {
		if err := s.producer.Publish(ctx, task); err != nil {
			log.Printf("ParseService: failed to publish task for request %s: %v", req.ID, err)
		} else {
			log.Printf("ParseService: published task for request %s, URL %s", req.ID, url)
		}
	} else {
		log.Printf("ParseService: producer is nil, task NOT published for request %s", req.ID)
	}

	return req, nil
}

func (s *ParseService) ParseBatch(
	ctx context.Context,
	urls []string,
	mediaTypeCodes []string,
	limit, offset, priority int,
	tokenID *int,
) (*entity.Request, error) {
	log.Printf("ParseBatch: called with %d URLs, tokenID=%v", len(urls), tokenID)

	req, err := s.requestService.CreateRequest(ctx, mediaTypeCodes, limit, offset, priority, tokenID)
	if err != nil {
		return nil, err
	}
	log.Printf("ParseBatch: request created %s", req.ID)

	for _, url := range urls {
		source, err := s.sourceService.GetOrCreateSource(ctx, url)
		if err != nil {
			log.Printf("ParseBatch: failed to get/create source for URL %s: %v", url, err)
			continue
		}
		log.Printf("ParseBatch: source created/found %d for URL %s", source.ID, url)

		_, err = s.requestService.AddSourceToRequest(ctx, req.ID, source.ID, limit)
		if err != nil {
			log.Printf("ParseBatch: failed to add source %d to request %s: %v", source.ID, req.ID, err)
			continue
		}

		task := &queue.ParseTask{
			RequestID:  req.ID,
			SourceID:   source.ID,
			URL:        url,
			MediaType:  nil,
			Limit:      limit,
			Priority:   priority,
			RetryCount: 0,
			MaxRetries: 3,
			CreatedAt:  req.CreatedAt,
		}
		if s.producer != nil {
			log.Printf("ParseBatch: publishing task for URL %s", url)
			if err := s.producer.Publish(ctx, task); err != nil {
				log.Printf("ParseBatch: failed to publish task for URL %s: %v", url, err)
			} else {
				log.Printf("ParseBatch: task published for URL %s", url)
			}
		} else {
			log.Printf("ParseBatch: producer is nil, task NOT published for URL %s", url)
		}
	}

	return req, nil
}

func (s *ParseService) ParseAllSources(
	ctx context.Context,
	mediaTypeCodes []string,
	limit, offset, priority int,
	tokenID *int,
) (*entity.Request, error) {
	sources, err := s.sourceRepo.GetActive(ctx)
	if err != nil {
		return nil, err
	}
	if len(sources) == 0 {
		return nil, errors.New("no active sources")
	}

	req, err := s.requestService.CreateRequest(ctx, mediaTypeCodes, limit, offset, priority, tokenID)
	if err != nil {
		return nil, err
	}

	for _, source := range sources {
		_, err := s.requestService.AddSourceToRequest(ctx, req.ID, source.ID, limit)
		if err != nil {
			return nil, err
		}

		task := &queue.ParseTask{
			RequestID:  req.ID,
			SourceID:   source.ID,
			URL:        source.BaseURL,
			MediaType:  nil,
			Limit:      limit,
			Priority:   priority,
			RetryCount: 0,
			MaxRetries: 3,
			CreatedAt:  req.CreatedAt,
		}
		if s.producer != nil {
			_ = s.producer.Publish(ctx, task)
		}
	}

	return req, nil
}

func (s *ParseService) ParseFirst(
	ctx context.Context,
	mediaTypeCodes []string,
	priority int,
	tokenID *int,
) (*entity.Request, error) {
	sources, err := s.sourceRepo.GetActive(ctx)
	if err != nil {
		return nil, err
	}
	if len(sources) == 0 {
		return nil, errors.New("no active sources")
	}

	req, err := s.requestService.CreateRequest(ctx, mediaTypeCodes, 1, 0, priority, tokenID)
	if err != nil {
		return nil, err
	}

	for _, source := range sources {
		_, err := s.requestService.AddSourceToRequest(ctx, req.ID, source.ID, 1)
		if err != nil {
			return nil, err
		}

		task := &queue.ParseTask{
			RequestID:  req.ID,
			SourceID:   source.ID,
			URL:        source.BaseURL,
			MediaType:  nil,
			Limit:      1,
			Priority:   priority,
			RetryCount: 0,
			MaxRetries: 3,
			CreatedAt:  req.CreatedAt,
		}
		if s.producer != nil {
			_ = s.producer.Publish(ctx, task)
		}
	}

	return req, nil
}

func (s *ParseService) ParseN(
	ctx context.Context,
	count int,
	mediaTypeCodes []string,
	offset, priority int,
	tokenID *int,
) (*entity.Request, error) {
	req, err := s.requestService.CreateRequest(ctx, mediaTypeCodes, count, offset, priority, tokenID)
	if err != nil {
		return nil, err
	}

	sources, err := s.sourceRepo.GetActive(ctx)
	if err != nil {
		return nil, err
	}

	if len(sources) > 0 {
		countPerSource := count / len(sources)
		if countPerSource < 1 {
			countPerSource = 1
		}

		for _, source := range sources {
			_, err := s.requestService.AddSourceToRequest(ctx, req.ID, source.ID, countPerSource)
			if err != nil {
				return nil, err
			}

			task := &queue.ParseTask{
				RequestID:  req.ID,
				SourceID:   source.ID,
				URL:        source.BaseURL,
				MediaType:  nil,
				Limit:      countPerSource,
				Priority:   priority,
				RetryCount: 0,
				MaxRetries: 3,
				CreatedAt:  req.CreatedAt,
			}
			if s.producer != nil {
				_ = s.producer.Publish(ctx, task)
			}
		}
	}

	return req, nil
}

func (s *ParseService) ParseSource(
	ctx context.Context,
	sourceID int,
	mediaTypeCodes []string,
	limit, offset, priority int,
	tokenID *int,
) (*entity.Request, error) {
	source, err := s.sourceRepo.GetByID(ctx, sourceID)
	if err != nil {
		return nil, err
	}
	if source == nil {
		return nil, errors.New("source not found")
	}

	req, err := s.requestService.CreateRequest(ctx, mediaTypeCodes, limit, offset, priority, tokenID)
	if err != nil {
		return nil, err
	}

	_, err = s.requestService.AddSourceToRequest(ctx, req.ID, sourceID, limit)
	if err != nil {
		return nil, err
	}

	task := &queue.ParseTask{
		RequestID:  req.ID,
		SourceID:   sourceID,
		URL:        source.BaseURL,
		MediaType:  nil,
		Limit:      limit,
		Priority:   priority,
		RetryCount: 0,
		MaxRetries: 3,
		CreatedAt:  req.CreatedAt,
	}
	if s.producer != nil {
		_ = s.producer.Publish(ctx, task)
	}

	return req, nil
}

func (s *ParseService) IsURLCached(ctx context.Context, url string) (bool, error) {
	return s.cacheRepo.IsURLCached(ctx, url)
}

func (s *ParseService) MarkURLAsParsed(ctx context.Context, url, hash string) error {
	return s.cacheRepo.SetURLCache(ctx, url, hash, 0)
}
