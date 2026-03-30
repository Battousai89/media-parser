package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/media-parser/backend/internal/model/dto"
	"github.com/media-parser/backend/internal/model/entity"
	"github.com/media-parser/backend/internal/repository"
)

var (
	ErrRequestNotFound = errors.New("request not found")
)

type RequestService struct {
	requestRepo       repository.RequestRepository
	requestSourceRepo repository.RequestSourceRepository
	dictRepo          repository.DictionaryRepository
}

func NewRequestService(
	requestRepo repository.RequestRepository,
	requestSourceRepo repository.RequestSourceRepository,
	dictRepo repository.DictionaryRepository,
) *RequestService {
	return &RequestService{
		requestRepo:       requestRepo,
		requestSourceRepo: requestSourceRepo,
		dictRepo:          dictRepo,
	}
}

func (s *RequestService) CreateRequest(
	ctx context.Context,
	mediaTypeCodes []string,
	limit, offset, priority int,
	tokenID *int,
) (*entity.Request, error) {
	status, err := s.dictRepo.GetRequestStatusByCode(ctx, entity.StatusPending)
	if err != nil {
		return nil, err
	}
	if status == nil {
		return nil, errors.New("pending status not found")
	}

	req := &entity.Request{
		ID:          uuid.New(),
		StatusID:    status.ID,
		LimitCount:  limit,
		OffsetCount: offset,
		Priority:    priority,
		MaxRetries:  3,
		RetryCount:  0,
		TokenID:     tokenID,
	}

	if len(mediaTypeCodes) > 0 {
		mediaTypeIDs := make([]int, 0, len(mediaTypeCodes))
		for _, code := range mediaTypeCodes {
			if code == "" || code == "all" {
				continue
			}
			mediaType, err := s.dictRepo.GetMediaTypeByCode(ctx, code)
			if err != nil {
				return nil, err
			}
			if mediaType != nil {
				mediaTypeIDs = append(mediaTypeIDs, mediaType.ID)
			}
		}
		if len(mediaTypeIDs) > 0 {
			req.MediaTypeIDs = mediaTypeIDs
		}
	}

	if err := s.requestRepo.Create(ctx, req); err != nil {
		return nil, err
	}

	if len(req.MediaTypeIDs) > 0 {
		if err := s.requestRepo.SetMediaTypeIDs(ctx, req.ID, req.MediaTypeIDs); err != nil {
			return nil, err
		}
	}

	return req, nil
}

func (s *RequestService) GetRequestByID(ctx context.Context, id uuid.UUID) (*entity.Request, error) {
	return s.requestRepo.GetByID(ctx, id)
}

func (s *RequestService) GetAllRequests(ctx context.Context, limit, offset int) ([]*entity.Request, error) {
	return s.requestRepo.GetAll(ctx, limit, offset)
}

func (s *RequestService) Count(ctx context.Context) (int, error) {
	return s.requestRepo.Count(ctx)
}

func (s *RequestService) GetRequestsByStatus(ctx context.Context, statusID int, limit, offset int) ([]*entity.Request, error) {
	return s.requestRepo.GetByStatus(ctx, statusID, limit, offset)
}

func (s *RequestService) UpdateRequestStatus(ctx context.Context, id uuid.UUID, statusCode string, errorMsg *string) error {
	status, err := s.dictRepo.GetRequestStatusByCode(ctx, statusCode)
	if err != nil {
		return err
	}
	if status == nil {
		return errors.New("invalid status code")
	}

	return s.requestRepo.UpdateStatus(ctx, id, status.ID, errorMsg)
}

func (s *RequestService) DeleteRequest(ctx context.Context, id uuid.UUID) error {
	return s.requestRepo.Delete(ctx, id)
}

func (s *RequestService) GetRequestMedia(ctx context.Context, requestID uuid.UUID, limit, offset int) ([]*entity.Media, error) {
	return s.requestRepo.GetRequestMedia(ctx, requestID, limit, offset)
}

func (s *RequestService) GetRequestMediaWithSources(ctx context.Context, requestID uuid.UUID, limit, offset int) ([]*dto.RequestMediaItem, error) {
	return s.requestRepo.GetRequestMediaWithSources(ctx, requestID, limit, offset)
}

func (s *RequestService) AddSourceToRequest(
	ctx context.Context,
	requestID uuid.UUID,
	sourceID int,
	mediaCount int,
) (*entity.RequestSource, error) {
	status, err := s.dictRepo.GetRequestStatusByCode(ctx, entity.StatusPending)
	if err != nil {
		return nil, err
	}
	if status == nil {
		return nil, errors.New("pending status not found")
	}

	rs := &entity.RequestSource{
		RequestID:   requestID,
		SourceID:    sourceID,
		StatusID:    status.ID,
		MediaCount:  mediaCount,
		ParsedCount: 0,
		RetryCount:  0,
		MaxRetries:  3,
	}

	if err := s.requestSourceRepo.Create(ctx, rs); err != nil {
		return nil, err
	}

	return rs, nil
}

func (s *RequestService) GetRequestSources(ctx context.Context, requestID uuid.UUID) ([]*entity.RequestSource, error) {
	return s.requestSourceRepo.GetByRequestID(ctx, requestID)
}

func (s *RequestService) UpdateRequestSourceStatus(
	ctx context.Context,
	requestID uuid.UUID,
	sourceID int,
	statusCode string,
	parsedCount int,
	errorMsg *string,
) error {
	status, err := s.dictRepo.GetRequestStatusByCode(ctx, statusCode)
	if err != nil {
		return err
	}
	if status == nil {
		return errors.New("invalid status code")
	}

	return s.requestSourceRepo.UpdateStatus(ctx, requestID, sourceID, status.ID, parsedCount, errorMsg)
}

func (s *RequestService) StartRequest(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	req, err := s.requestRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if req == nil {
		return ErrRequestNotFound
	}

	req.StartedAt = &now
	return s.requestRepo.Update(ctx, req)
}

func (s *RequestService) CompleteRequest(ctx context.Context, id uuid.UUID, statusCode string, errorMsg *string) error {
	now := time.Now()
	req, err := s.requestRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if req == nil {
		return ErrRequestNotFound
	}

	status, err := s.dictRepo.GetRequestStatusByCode(ctx, statusCode)
	if err != nil {
		return err
	}
	if status == nil {
		return errors.New("invalid status code")
	}

	req.StatusID = status.ID
	req.CompletedAt = &now
	req.ErrorMessage = errorMsg

	return s.requestRepo.Update(ctx, req)
}
