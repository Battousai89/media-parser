package queue

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/media-parser/backend/internal/config"
	"github.com/media-parser/backend/internal/model/entity"
	"github.com/media-parser/backend/internal/parser"
	"github.com/media-parser/backend/internal/parser/browser"
	"github.com/media-parser/backend/internal/parser/youtube"
	"github.com/media-parser/backend/internal/repository/minio"
	"github.com/media-parser/backend/internal/repository"
	"github.com/wader/goutubedl"
)

type ParseWorker struct {
	sourceRepo        repository.SourceRepository
	patternRepo       repository.PatternRepository
	mediaRepo         repository.MediaRepository
	sourceMediaRepo   repository.SourceMediaRepository
	requestRepo       repository.RequestRepository
	requestSourceRepo repository.RequestSourceRepository
	cacheRepo         repository.RedisCacheRepository
	dictRepo          repository.DictionaryRepository
	extRepo           repository.MediaTypeExtensionRepository
	httpClient        *parser.HTTPClient
	browserDriver     *browser.Driver
	youtubeParser     *youtube.Parser
	minioClient       *minio.Client
	detector          *parser.MediaTypeDetector
	matcher           *parser.PatternMatcher
	robotsChecker     *parser.RobotsChecker
	statusCodes       map[string]int32
	sourceStatusCodes map[string]int32
	mediaTypeCodes    map[string]int32
	cacheTTL          time.Duration
	mu                sync.RWMutex
	youtubeEnabled    bool
}

func NewParseWorker(
	sourceRepo repository.SourceRepository,
	patternRepo repository.PatternRepository,
	mediaRepo repository.MediaRepository,
	sourceMediaRepo repository.SourceMediaRepository,
	requestRepo repository.RequestRepository,
	requestSourceRepo repository.RequestSourceRepository,
	cacheRepo repository.RedisCacheRepository,
	dictRepo repository.DictionaryRepository,
	extRepo repository.MediaTypeExtensionRepository,
	minioClient *minio.Client,
	cfg *config.ParserConfig,
) *ParseWorker {
	cacheTTL := time.Duration(cfg.CacheTTLHours) * time.Hour

	browserDriver := browser.NewDriver(browser.Config{
		Headless:        cfg.BrowserHeadless,
		PageLoadTimeout: time.Duration(cfg.PageLoadTimeout) * time.Second,
		WindowWidth:     1920,
		WindowHeight:    1080,
	})

	youtubeParser := youtube.NewParser(time.Duration(cfg.YouTubeTimeout) * time.Second)

	w := &ParseWorker{
		sourceRepo:        sourceRepo,
		patternRepo:       patternRepo,
		mediaRepo:         mediaRepo,
		sourceMediaRepo:   sourceMediaRepo,
		requestRepo:       requestRepo,
		requestSourceRepo: requestSourceRepo,
		cacheRepo:         cacheRepo,
		dictRepo:          dictRepo,
		extRepo:           extRepo,
		httpClient:        parser.NewHTTPClient(parser.HTTPClientConfig{Timeout: 30 * time.Second}),
		browserDriver:     browserDriver,
		youtubeParser:     youtubeParser,
		minioClient:       minioClient,
		detector:          parser.NewMediaTypeDetector(),
		matcher:           parser.NewPatternMatcher(),
		robotsChecker:     parser.NewRobotsChecker(cacheTTL, cfg.ParserUserAgent, cfg.IgnoreRobotsTxt),
		statusCodes:       make(map[string]int32),
		sourceStatusCodes: make(map[string]int32),
		mediaTypeCodes:    make(map[string]int32),
		cacheTTL:          cacheTTL,
		youtubeEnabled:    cfg.YouTubeEnabled,
	}

	w.loadStatusCodes()
	w.loadSourceStatusCodes()
	w.loadMediaTypeCodes()
	return w
}

func (w *ParseWorker) loadStatusCodes() {
	w.mu.Lock()
	defer w.mu.Unlock()

	statuses, err := w.dictRepo.GetRequestStatuses(context.Background())
	if err != nil {
		log.Printf("ParseWorker: failed to load status codes: %v", err)
		return
	}
	for _, s := range statuses {
		w.statusCodes[s.Code] = int32(s.ID)
	}
}

func (w *ParseWorker) loadSourceStatusCodes() {
	w.mu.Lock()
	defer w.mu.Unlock()

	statuses, err := w.dictRepo.GetSourceStatuses(context.Background())
	if err != nil {
		log.Printf("ParseWorker: failed to load source status codes: %v", err)
		return
	}
	for _, s := range statuses {
		w.sourceStatusCodes[s.Code] = int32(s.ID)
	}
}

func (w *ParseWorker) loadMediaTypeCodes() {
	w.mu.Lock()
	defer w.mu.Unlock()

	mediaTypes, err := w.dictRepo.GetMediaTypes(context.Background())
	if err != nil {
		return
	}
	for _, t := range mediaTypes {
		w.mediaTypeCodes[t.Code] = int32(t.ID)
	}
}

func (w *ParseWorker) statusCode(code string) int {
	w.mu.RLock()
	id, ok := w.statusCodes[code]
	w.mu.RUnlock()

	if ok {
		return int(id)
	}

	w.loadStatusCodes()

	w.mu.RLock()
	defer w.mu.RUnlock()
	if id, ok := w.statusCodes[code]; ok {
		return int(id)
	}
	return 1
}

func (w *ParseWorker) mediaTypeID(code string) int {
	w.mu.RLock()
	id, ok := w.mediaTypeCodes[code]
	w.mu.RUnlock()

	if ok {
		return int(id)
	}

	w.loadMediaTypeCodes()

	w.mu.RLock()
	defer w.mu.RUnlock()
	if id, ok := w.mediaTypeCodes[code]; ok {
		return int(id)
	}
	return 1
}

func (w *ParseWorker) Handle(ctx context.Context, task *ParseTask) error {
	log.Printf("ParseWorker: handling task for request %s, source %d, URL %s", task.RequestID, task.SourceID, task.URL)

	processingStatus := w.statusCode("processing")
	if err := w.requestRepo.UpdateStatus(ctx, task.RequestID, processingStatus, nil); err != nil {
		log.Printf("ParseWorker: failed to update request status: %v", err)
		return fmt.Errorf("update request status: %w", err)
	}

	if err := w.requestSourceRepo.UpdateStatus(ctx, task.RequestID, task.SourceID, processingStatus, 0, nil); err != nil {
		log.Printf("ParseWorker: failed to update request source status: %v", err)
		return fmt.Errorf("update request source status: %w", err)
	}

	cached, err := w.cacheRepo.GetURLCache(ctx, task.URL)
	if err == nil && cached != nil && time.Since(cached.ParsedAt) < w.cacheTTL {
		existingReqSource, _ := w.requestSourceRepo.GetByRequestIDAndSourceID(ctx, task.RequestID, task.SourceID)
		if existingReqSource != nil && existingReqSource.StatusID == w.statusCode("failed") {
			log.Printf("ParseWorker: URL %s is cached but previous request failed, re-parsing", task.URL)
		} else {
			log.Printf("ParseWorker: URL %s is cached, loading media from DB", task.URL)

			// Находим все медиа для этого источника и создаём связи с новым запросом
			sourceMedias, err := w.sourceMediaRepo.GetBySourceID(ctx, task.SourceID, 1000, 0)
			if err != nil {
				log.Printf("ParseWorker: GetBySourceID error: %v, re-parsing", err)
			} else if len(sourceMedias) > 0 {
				log.Printf("ParseWorker: found %d cached media items for source %d", len(sourceMedias), task.SourceID)
				parsedCount := 0
				for _, sm := range sourceMedias {
					// Создаём связь с новым запросом
					newSM := &entity.SourceMedia{
						SourceID:  task.SourceID,
						MediaID:   sm.MediaID,
						RequestID: &task.RequestID,
					}
					_ = w.sourceMediaRepo.Create(ctx, newSM)
					parsedCount++
				}
				_ = w.cacheRepo.SetURLCache(ctx, task.URL, w.hashURL(task.URL), w.cacheTTL)
				_ = w.sourceRepo.UpdateStatus(ctx, task.SourceID, w.sourceStatusCode("active"))
				_ = w.requestSourceRepo.UpdateStatus(ctx, task.RequestID, task.SourceID, w.statusCode("completed"), parsedCount, nil)
				w.checkRequestCompletion(ctx, task.RequestID)
				return nil
			} else {
				log.Printf("ParseWorker: no cached media found for source, re-parsing URL %s", task.URL)
			}
		}
	}

	log.Printf("ParseWorker: starting to parse URL %s", task.URL)

	// Проверяем, является ли URL YouTube и включен ли YouTube парсер
	if w.youtubeEnabled && youtube.IsYouTubeURL(task.URL) {
		log.Printf("ParseWorker: detected YouTube URL %s, using yt-dlp parser", task.URL)
		return w.handleYouTubeURL(ctx, task)
	}

	canFetch, err := w.robotsChecker.CanFetch(task.URL)
	if err != nil {
		log.Printf("ParseWorker: robots check error: %v", err)
	}
	if err == nil && !canFetch {
		errMsg := "blocked by robots.txt"
		log.Printf("ParseWorker: blocked by robots.txt")
		_ = w.sourceRepo.UpdateStatus(ctx, task.SourceID, w.sourceStatusCode("blocked"))
		_ = w.requestSourceRepo.UpdateStatus(ctx, task.RequestID, task.SourceID, w.statusCode("failed"), 0, &errMsg)
		return fmt.Errorf("blocked by robots.txt")
	}

	log.Printf("ParseWorker: fetching URL %s via chromedp", task.URL)
	html, err := w.browserDriver.GetPageSource(ctx, task.URL)
	if err != nil {
		errMsg := err.Error()
		log.Printf("ParseWorker: chromedp error: %v", err)
		_ = w.sourceRepo.UpdateStatus(ctx, task.SourceID, w.sourceStatusCode("error"))
		_ = w.requestSourceRepo.UpdateStatus(ctx, task.RequestID, task.SourceID, w.statusCode("failed"), 0, &errMsg)
		return fmt.Errorf("fetch page via chromedp: %w", err)
	}

	log.Printf("ParseWorker: fetched %d bytes via chromedp", len(html))

	var patterns []*entity.Pattern
	if task.SourceID > 0 {
		patterns, err = w.patternRepo.GetAll(ctx)
		if err != nil {
			log.Printf("ParseWorker: get patterns error: %v", err)
			return fmt.Errorf("get patterns: %w", err)
		}
	}

	if len(patterns) == 0 {
		log.Printf("ParseWorker: no patterns, using defaults")
		patterns = w.getDefaultPatterns()
	}

	_ = w.matcher.CompileAll(patterns)

	mediaURLs := w.matcher.MatchAll(html, patterns)
	log.Printf("ParseWorker: matcher found %d media types", len(mediaURLs))

	parsedCount := 0
	for mediaTypeID, urls := range mediaURLs {
		log.Printf("ParseWorker: media type ID %d has %d URLs", mediaTypeID, len(urls))
		for _, u := range urls {
			if parsedCount >= task.Limit {
				break
			}

			log.Printf("ParseWorker: raw URL: %s", u)
			fullURL := w.normalizeURL(u, task.URL)
			log.Printf("ParseWorker: normalized URL: %s -> %s", u, fullURL)
			if fullURL == "" {
				log.Printf("ParseWorker: skipped empty normalized URL")
				continue
			}

			hash := w.hashURL(fullURL)
			existing, _ := w.mediaRepo.GetByHash(ctx, hash)
			if existing != nil {
				log.Printf("ParseWorker: media already exists by hash, linking to request")
				// Создаём связь в source_media для этого запроса
				sm := &entity.SourceMedia{
					SourceID:  task.SourceID,
					MediaID:   existing.ID,
					RequestID: &task.RequestID,
				}
				_ = w.sourceMediaRepo.Create(ctx, sm)
				parsedCount++
				continue
			}

			existingByURL, _ := w.mediaRepo.GetByURL(ctx, fullURL)
			if existingByURL != nil {
				log.Printf("ParseWorker: media already exists by URL, linking to request")
				// Создаём связь в source_media для этого запроса
				sm := &entity.SourceMedia{
					SourceID:  task.SourceID,
					MediaID:   existingByURL.ID,
					RequestID: &task.RequestID,
				}
				_ = w.sourceMediaRepo.Create(ctx, sm)
				parsedCount++
				continue
			}

			// Определить расширение из URL
			ext := strings.ToLower(filepath.Ext(fullURL))
			if ext == "" {
				ext = ".bin"
			}

			// Найти media_type_id по расширению через extRepo
			mediaTypeID, err := w.extRepo.GetMediaTypeByExtension(ctx, ext)
			if err != nil || mediaTypeID == nil {
				// Fallback: определить по детектору
				mediaType := w.detector.DetectByURL(fullURL)
				mtid := w.getMediaTypeID(mediaType)
				mediaTypeID = &mtid
			}
			log.Printf("ParseWorker: media type_id=%d, extension=%s", *mediaTypeID, ext)

			// Скачать в Minio
			storageKey, err := w.downloadToMinio(ctx, fullURL, *mediaTypeID, "")
			if err != nil {
				log.Printf("ParseWorker: failed to download to minio: %v", err)
				continue
			}

			// Создать запись в media
			media := &entity.Media{
				ID:          uuid.New(),
				URL:         fullURL,
				MediaTypeID: *mediaTypeID,
				StoragePath: &storageKey,
				Hash:        &hash,
				Available:   true,
			}

			if err := w.mediaRepo.Create(ctx, media); err != nil {
				log.Printf("ParseWorker: failed to create media: %v", err)
				continue
			}

			sm := &entity.SourceMedia{
				SourceID:  task.SourceID,
				MediaID:   media.ID,
				RequestID: &task.RequestID,
			}
			if err := w.sourceMediaRepo.Create(ctx, sm); err != nil {
				log.Printf("ParseWorker: failed to create source_media: %v", err)
				continue
			}

			parsedCount++
			log.Printf("ParseWorker: saved media %s to minio %s", media.ID, storageKey)
		}

		if parsedCount >= task.Limit {
			break
		}
	}

	_ = w.cacheRepo.SetURLCache(ctx, task.URL, w.hashURL(task.URL), w.cacheTTL)

	log.Printf("ParseWorker: parsed %d media items", parsedCount)

	_ = w.sourceRepo.UpdateStatus(ctx, task.SourceID, w.sourceStatusCode("active"))

	// parsedCount == 0 это НЕ ошибка, это completed успешно
	sourceStatus := w.statusCode("completed")
	_ = w.requestSourceRepo.UpdateStatus(ctx, task.RequestID, task.SourceID, sourceStatus, parsedCount, nil)

	// checkRequestCompletion вызывается в worker когда все retry исчерпаны ИЛИ здесь если нет ошибки
	w.checkRequestCompletion(ctx, task.RequestID)

	return nil
}

func (w *ParseWorker) checkRequestCompletion(ctx context.Context, requestID uuid.UUID) {
	sources, _ := w.requestSourceRepo.GetByRequestID(ctx, requestID)

	completedStatus := w.statusCode("completed")
	failedStatus := w.statusCode("failed")
	processingStatus := w.statusCode("processing")

	completedCount := 0
	failedCount := 0
	processingCount := 0

	for _, s := range sources {
		if s.StatusID == processingStatus {
			processingCount++
		} else if s.StatusID == completedStatus {
			completedCount++
		} else if s.StatusID == failedStatus {
			failedCount++
		}
	}

	if processingCount > 0 {
		return  // ещё не все обработаны
	}

	var statusCode string
	if failedCount > 0 && completedCount > 0 {
		statusCode = "partial"   // часть успех, часть ошибка
	} else if failedCount > 0 {
		statusCode = "failed"    // все failed
	} else {
		statusCode = "completed" // все completed
	}

	log.Printf("ParseWorker: all sources completed, updating request status to %s (completed=%d, failed=%d)", statusCode, completedCount, failedCount)
	err := w.requestRepo.UpdateStatus(ctx, requestID, w.statusCode(statusCode), nil)
	if err != nil {
		log.Printf("ParseWorker: failed to update request status: %v", err)
	}
}

// handleYouTubeURL обрабатывает YouTube URL через goutubedl
func (w *ParseWorker) handleYouTubeURL(ctx context.Context, task *ParseTask) error {
	// Нормализуем URL для goutubedl (music.youtube.com -> www.youtube.com, убираем list параметр)
	normalizedURL := strings.Replace(task.URL, "music.youtube.com", "www.youtube.com", 1)
	if idx := strings.Index(normalizedURL, "&list="); idx >= 0 {
		normalizedURL = normalizedURL[:idx]
	}
	log.Printf("ParseWorker: processing YouTube URL %s via goutubedl (normalized: %s)", task.URL, normalizedURL)

	// Получаем информацию через goutubedl
	result, err := goutubedl.New(ctx, normalizedURL, goutubedl.Options{})
	if err != nil {
		errMsg := fmt.Sprintf("goutubedl error: %v", err)
		log.Printf("ParseWorker: goutubedl error for URL %s: %v", task.URL, err)

		_ = w.sourceRepo.UpdateStatus(ctx, task.SourceID, w.sourceStatusCode("error"))
		_ = w.requestSourceRepo.UpdateStatus(ctx, task.RequestID, task.SourceID, w.statusCode("failed"), 0, &errMsg)
		return fmt.Errorf("goutubedl parse: %w", err)
	}

	log.Printf("ParseWorker: goutubedl extracted: title=%q, duration=%v", result.Info.Title, result.Info.Duration)

	parsedCount := 0
	ext := ".webm"

	// Найти media_type_id
	mediaTypeID, err := w.extRepo.GetMediaTypeByExtension(ctx, ext)
	if err != nil || mediaTypeID == nil {
		mtid := w.mediaTypeID("video_audio")
		mediaTypeID = &mtid
	}

	// Скачиваем через goutubedl напрямую в MinIO
	log.Printf("ParseWorker: downloading audio via goutubedl")
	
	downloadResult, err := result.Download(ctx, "bestaudio")
	if err != nil {
		errMsg := fmt.Sprintf("goutubedl download error: %v", err)
		log.Printf("ParseWorker: goutubedl download error: %v", err)
		_ = w.sourceRepo.UpdateStatus(ctx, task.SourceID, w.sourceStatusCode("error"))
		_ = w.requestSourceRepo.UpdateStatus(ctx, task.RequestID, task.SourceID, w.statusCode("failed"), 0, &errMsg)
		return fmt.Errorf("goutubedl download: %w", err)
	}
	defer downloadResult.Close()

	// Читаем данные
	body, err := io.ReadAll(downloadResult)
	if err != nil {
		return fmt.Errorf("read download result: %w", err)
	}

	log.Printf("ParseWorker: downloaded %d bytes", len(body))

	// Вычисляем hash
	hash := sha256.Sum256(body)
	hashHex := hex.EncodeToString(hash[:])
	storageKey := fmt.Sprintf("media/%d/%s%s", *mediaTypeID, hashHex, ext)

	// Проверяем существует ли файл в БД
	existing, _ := w.mediaRepo.GetByHash(ctx, hashHex)
	if existing != nil {
		if existing.StoragePath == nil || *existing.StoragePath == "" {
			log.Printf("ParseWorker: media exists but storage_path is empty, uploading to minio")
			err = w.minioClient.Upload(ctx, storageKey, bytes.NewReader(body), int64(len(body)), "audio/webm")
			if err != nil {
				log.Printf("ParseWorker: failed to upload to minio: %v", err)
			} else {
				err = w.mediaRepo.Update(ctx, &entity.Media{
					ID:          existing.ID,
					StoragePath: &storageKey,
				})
				if err != nil {
					log.Printf("ParseWorker: failed to update media: %v", err)
				} else {
					log.Printf("ParseWorker: updated storage path for media %s to %s", existing.ID, storageKey)
				}
			}
		}
		sm := &entity.SourceMedia{
			SourceID:  task.SourceID,
			MediaID:   existing.ID,
			RequestID: &task.RequestID,
		}
		if err := w.sourceMediaRepo.Create(ctx, sm); err == nil {
			parsedCount++
		}
		return nil
	}

	// Upload в MinIO перед созданием записи в БД
	log.Printf("ParseWorker: uploading to minio with key %s", storageKey)
	err = w.minioClient.Upload(ctx, storageKey, bytes.NewReader(body), int64(len(body)), "audio/webm")
	if err != nil {
		log.Printf("ParseWorker: failed to upload to minio: %v", err)
		return fmt.Errorf("upload to minio: %w", err)
	}
	log.Printf("ParseWorker: uploaded to minio successfully")

	// Создаём новую запись
	media := &entity.Media{
		ID:          uuid.New(),
		URL:         task.URL,
		MediaTypeID: *mediaTypeID,
		StoragePath: &storageKey,
		Hash:        &hashHex,
		Available:   true,
		Title:       &result.Info.Title,
	}

	if err := w.mediaRepo.Create(ctx, media); err == nil {
		sm := &entity.SourceMedia{
			SourceID:  task.SourceID,
			MediaID:   media.ID,
			RequestID: &task.RequestID,
		}
		if err := w.sourceMediaRepo.Create(ctx, sm); err == nil {
			parsedCount++
			log.Printf("ParseWorker: saved audio media %s to minio %s", media.ID, storageKey)
		}
	}

	_ = w.cacheRepo.SetURLCache(ctx, task.URL, w.hashURL(task.URL), w.cacheTTL)

	log.Printf("ParseWorker: YouTube parsed %d media items", parsedCount)

	_ = w.sourceRepo.UpdateStatus(ctx, task.SourceID, w.sourceStatusCode("active"))
	sourceStatus := w.statusCode("completed")
	_ = w.requestSourceRepo.UpdateStatus(ctx, task.RequestID, task.SourceID, sourceStatus, parsedCount, nil)

	w.checkRequestCompletion(ctx, task.RequestID)

	return nil
}

func (w *ParseWorker) downloadToMinio(ctx context.Context, url string, mediaTypeID int, ext string) (string, error) {
	log.Printf("ParseWorker: downloadToMinio called for media type ID %d, URL: %s", mediaTypeID, url)

	var reader io.Reader
	var contentType string
	var err error

	// Для YouTube URL используем goutubedl
	if youtube.IsYouTubeURL(url) {
		log.Printf("ParseWorker: using goutubedl for YouTube URL")
		
		result, err := goutubedl.New(ctx, url, goutubedl.Options{})
		if err != nil {
			return "", fmt.Errorf("goutubedl new: %w", err)
		}

		downloadResult, err := result.Download(ctx, "bestaudio")
		if err != nil {
			return "", fmt.Errorf("goutubedl download: %w", err)
		}
		defer downloadResult.Close()

		reader = downloadResult
		contentType = "audio/webm"
		if ext == "" {
			ext = ".webm"
		}
		log.Printf("ParseWorker: goutubedl download started, extension: %s", ext)
	} else {
		// Для обычных URL используем HTTP клиент
		client := &http.Client{
			Timeout: 30 * time.Second,
		}

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return "", fmt.Errorf("create request: %w", err)
		}
		req.Header.Set("User-Agent", "MediaParserBot/1.0")

		resp, err := client.Do(req)
		if err != nil {
			return "", fmt.Errorf("fetch url: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("unexpected status: %s", resp.Status)
		}

		reader = resp.Body
		contentType = resp.Header.Get("Content-Type")
		
		// Определяем расширение из Content-Type если не передано
		if ext == "" && contentType != "" {
			exts, err := mime.ExtensionsByType(contentType)
			if err == nil && len(exts) > 0 {
				ext = exts[0]
			}
		}
		if ext == "" {
			ext = ".bin"
		}
		log.Printf("ParseWorker: HTTP download, content-type: %s, extension: %s", contentType, ext)
	}

	// Читаем данные
	body, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}

	log.Printf("ParseWorker: downloaded %d bytes", len(body))

	// Вычисляем hash содержимого файла
	hash := sha256.Sum256(body)
	hashHex := hex.EncodeToString(hash[:])

	// Используем hash как ключ для дедупликации
	storageKey := fmt.Sprintf("media/%d/%s%s", mediaTypeID, hashHex, ext)
	log.Printf("ParseWorker: file hash %s, storage key %s", hashHex[:16], storageKey)

	// Проверяем существует ли файл в Minio
	exists, err := w.minioClient.Exists(ctx, storageKey)
	if err != nil {
		log.Printf("ParseWorker: error checking file existence: %v", err)
	}
	if exists {
		log.Printf("ParseWorker: file already exists in minio %s", storageKey)
		return storageKey, nil
	}

	log.Printf("ParseWorker: uploading to minio with key %s", storageKey)
	err = w.minioClient.Upload(ctx, storageKey, bytes.NewReader(body), int64(len(body)), contentType)
	if err != nil {
		return "", fmt.Errorf("upload to minio: %w", err)
	}

	log.Printf("ParseWorker: uploaded to minio successfully")
	return storageKey, nil
}

func (w *ParseWorker) getDefaultPatterns() []*entity.Pattern {
	return []*entity.Pattern{
		{ID: 1, Name: "Images", Regex: `src="([^"]+\.(?:jpg|jpeg|png|webp|gif))"`, MediaTypeID: 1},
		{ID: 2, Name: "Videos", Regex: `src="([^"]+\.(?:mp4|webm))"`, MediaTypeID: 2},
		{ID: 3, Name: "Audio", Regex: `src="([^"]+\.(?:mp3|ogg))"`, MediaTypeID: 3},
	}
}

func (w *ParseWorker) normalizeURL(urlStr, baseURL string) string {
	if urlStr == "" {
		return ""
	}

	if len(urlStr) > 10 && (urlStr[:5] == "data:" || urlStr[:11] == "javascript:") {
		return ""
	}

	if len(urlStr) > 4 && urlStr[:4] == "http" {
		return urlStr
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}

	rel, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}

	resolved := base.ResolveReference(rel)

	if resolved.Scheme == "" {
		resolved.Scheme = base.Scheme
	}
	if resolved.Host == "" {
		resolved.Host = base.Host
	}

	return resolved.String()
}

func (w *ParseWorker) hashURL(url string) string {
	hash := sha256.Sum256([]byte(url))
	return hex.EncodeToString(hash[:])
}

func (w *ParseWorker) getMediaTypeID(mediaType string) int {
	return w.mediaTypeID(mediaType)
}

func (w *ParseWorker) sourceStatusCode(code string) int {
	w.mu.RLock()
	id, ok := w.sourceStatusCodes[code]
	w.mu.RUnlock()

	if ok {
		return int(id)
	}

	w.loadSourceStatusCodes()

	w.mu.RLock()
	defer w.mu.RUnlock()
	if id, ok := w.sourceStatusCodes[code]; ok {
		return int(id)
	}
	return 1
}

// getAudioFormatExtension извлекает расширение из лучшего аудио формата YouTube
func (w *ParseWorker) getAudioFormatExtension(info *youtube.VideoInfo) string {
	// Ищем лучший аудио формат
	for _, format := range info.Formats {
		if format.ACodec != "none" && format.ACodec != "" && format.Ext != "" {
			// Предпочитаем m4a как наиболее совместимый
			if format.Ext == "m4a" {
				return ".m4a"
			}
			// Возвращаем первое найденное аудио расширение
			return "." + format.Ext
		}
	}
	// Default для audio
	return ".m4a"
}
