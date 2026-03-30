package service

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/media-parser/backend/internal/model/entity"
	"github.com/media-parser/backend/internal/parser/youtube"
	"github.com/media-parser/backend/internal/repository"
	"github.com/media-parser/backend/internal/repository/minio"
)

type DownloadResult struct {
	StoragePath string
	FileSize    int64
	MimeType    string
}

type DownloadService struct {
	mediaRepo       repository.MediaRepository
	sourceMediaRepo repository.SourceMediaRepository
	minioClient     *minio.Client
	youtubeParser   *youtube.Parser
	downloadTimeout time.Duration
}

func NewDownloadService(
	mediaRepo repository.MediaRepository,
	sourceMediaRepo repository.SourceMediaRepository,
	minioClient *minio.Client,
	downloadTimeout time.Duration,
) *DownloadService {
	return &DownloadService{
		mediaRepo:       mediaRepo,
		sourceMediaRepo: sourceMediaRepo,
		minioClient:     minioClient,
		youtubeParser:   youtube.NewParser(30 * time.Second),
		downloadTimeout: downloadTimeout,
	}
}

func (s *DownloadService) GetMediaByID(ctx context.Context, id uuid.UUID) (*entity.Media, error) {
	return s.mediaRepo.GetByID(ctx, id)
}

func (s *DownloadService) GetFreshYouTubeURL(ctx context.Context, videoID string) (string, error) {
	return s.youtubeParser.GetFreshAudioURL(ctx, videoID)
}

// GetFileReader возвращает io.Reader для файла из Minio
func (s *DownloadService) GetFileReader(ctx context.Context, storagePath string) (io.Reader, error) {
	return s.minioClient.Download(ctx, storagePath)
}

func (s *DownloadService) DownloadFromMinio(ctx context.Context, storagePath string) (*DownloadResult, error) {
	reader, err := s.minioClient.Download(ctx, storagePath)
	if err != nil {
		return nil, fmt.Errorf("download from minio: %w", err)
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	return &DownloadResult{
		StoragePath: storagePath,
		FileSize:    int64(len(body)),
		MimeType:    "application/octet-stream",
	}, nil
}

func (s *DownloadService) GetPresignedURL(ctx context.Context, storagePath string) (string, error) {
	return s.minioClient.GetPresignedURL(ctx, storagePath, 7*24*time.Hour)
}
