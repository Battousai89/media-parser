package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/media-parser/backend/internal/model/entity"
	"github.com/media-parser/backend/internal/repository"
)

type SourceService struct {
	sourceRepo  repository.SourceRepository
	patternRepo repository.PatternRepository
	dictRepo    repository.DictionaryRepository
}

func NewSourceService(
	sourceRepo repository.SourceRepository,
	patternRepo repository.PatternRepository,
	dictRepo repository.DictionaryRepository,
) *SourceService {
	return &SourceService{
		sourceRepo:  sourceRepo,
		patternRepo: patternRepo,
		dictRepo:    dictRepo,
	}
}

func (s *SourceService) GetAll(ctx context.Context) ([]*entity.Source, error) {
	return s.sourceRepo.GetAll(ctx)
}

func (s *SourceService) GetActive(ctx context.Context) ([]*entity.Source, error) {
	return s.sourceRepo.GetActive(ctx)
}

func (s *SourceService) GetByID(ctx context.Context, id int) (*entity.Source, error) {
	return s.sourceRepo.GetByID(ctx, id)
}

func (s *SourceService) Create(ctx context.Context, name, baseURL string, statusID *int) (*entity.Source, error) {
	sid := 1
	if statusID != nil {
		sid = *statusID
	} else {
		status, err := s.dictRepo.GetSourceStatusByCode(ctx, entity.SourceStatusActive)
		if err == nil && status != nil {
			sid = status.ID
		}
	}

	source := &entity.Source{
		Name:     name,
		BaseURL:  baseURL,
		StatusID: sid,
	}

	if err := s.sourceRepo.Create(ctx, source); err != nil {
		return nil, fmt.Errorf("create source: %w", err)
	}

	return source, nil
}

func (s *SourceService) Update(ctx context.Context, id int, name *string, baseURL *string, statusID *int) (*entity.Source, error) {
	source, err := s.sourceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if source == nil {
		return nil, ErrSourceNotFound
	}

	if name != nil {
		source.Name = *name
	}
	if baseURL != nil {
		source.BaseURL = *baseURL
	}
	if statusID != nil {
		source.StatusID = *statusID
	}

	if err := s.sourceRepo.Update(ctx, source); err != nil {
		return nil, err
	}

	return source, nil
}

func (s *SourceService) Delete(ctx context.Context, id int) error {
	return s.sourceRepo.Delete(ctx, id)
}

func (s *SourceService) GetOrCreateSource(ctx context.Context, url string) (*entity.Source, error) {
	source, err := s.sourceRepo.GetByURL(ctx, url)
	if err != nil {
		return nil, err
	}
	if source != nil {
		statusID, err := s.checkSourceStatus(ctx, url)
		if err != nil {
			statusID, _ = s.getSourceStatusID(ctx, entity.SourceStatusError)
		}
		if statusID != source.StatusID {
			source.StatusID = statusID
			_ = s.sourceRepo.Update(ctx, source)
		}
		return source, nil
	}

	statusCode, err := s.checkSourceStatus(ctx, url)
	if err != nil {
		statusCode, _ = s.getSourceStatusID(ctx, entity.SourceStatusError)
	}

	name := extractDomain(url)
	if len(name) > 255 {
		name = name[:255]
	}
	return s.Create(ctx, name, url, &statusCode)
}

func extractDomain(url string) string {
	domain := url
	if len(domain) > 4 && domain[:4] == "http" {
		if domain[:5] == "https" {
			domain = domain[8:]
		} else {
			domain = domain[7:]
		}
	}
	if idx := indexOf(domain, '/'); idx != -1 {
		domain = domain[:idx]
	}
	if idx := indexOf(domain, '?'); idx != -1 {
		domain = domain[:idx]
	}
	if len(domain) > 4 && domain[:4] == "www." {
		domain = domain[4:]
	}
	return domain
}

func indexOf(s string, char byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == char {
			return i
		}
	}
	return -1
}

func (s *SourceService) checkSourceStatus(ctx context.Context, url string) (int, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return s.getSourceStatusID(ctx, entity.SourceStatusError)
	}
	req.Header.Set("User-Agent", "MediaParserBot/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return s.getSourceStatusID(ctx, entity.SourceStatusError)
	}
	defer resp.Body.Close()

	statusCode := entity.SourceStatusActive
	if resp.StatusCode >= 400 {
		if resp.StatusCode == 403 || resp.StatusCode == 429 {
			statusCode = entity.SourceStatusBlocked
		} else {
			statusCode = entity.SourceStatusError
		}
	}

	return s.getSourceStatusID(ctx, statusCode)
}

func (s *SourceService) getSourceStatusID(ctx context.Context, code string) (int, error) {
	status, err := s.dictRepo.GetSourceStatusByCode(ctx, code)
	if err != nil {
		return 0, err
	}
	if status == nil {
		return 0, errors.New("source status not found: " + code)
	}
	return status.ID, nil
}

var ErrSourceNotFound = errors.New("source not found")
