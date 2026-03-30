package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"

	"github.com/media-parser/backend/internal/config"
	"github.com/media-parser/backend/internal/handler"
	"github.com/media-parser/backend/internal/handler/middleware"
	"github.com/media-parser/backend/internal/queue"
	"github.com/media-parser/backend/internal/repository/minio"
	"github.com/media-parser/backend/internal/repository/postgres"
	redisCache "github.com/media-parser/backend/internal/repository/redis"
	"github.com/media-parser/backend/internal/service"
)

type APITestSuite struct {
	suite.Suite
	dbPool    *pgxpool.Pool
	redis     *redis.Client
	router    *gin.Engine
	ctx       context.Context
	cancel    context.CancelFunc
	testID    string
	testToken string
}

func (s *APITestSuite) SetupSuite() {
	s.ctx, s.cancel = context.WithTimeout(context.Background(), 5*time.Minute)
	s.testID = fmt.Sprintf("test_%s", uuid.New().String()[:8])
	s.testToken = fmt.Sprintf("test_token_%s", uuid.New().String()[:8])

	dbURL := "postgres://media_parser_user:media_parser_password@localhost:5432/media_parser?sslmode=disable"
	pool, err := pgxpool.New(s.ctx, dbURL)
	s.Require().NoError(err)
	s.dbPool = pool

	s.redis = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "media_parser_password",
		DB:       0,
	})

	s.createTestToken()

	db := &postgres.PostgresDB{Pool: s.dbPool}
	sourceRepo := postgres.NewSourceRepository(db)
	patternRepo := postgres.NewPatternRepository(db)
	requestRepo := postgres.NewRequestRepository(db)
	requestSourceRepo := postgres.NewRequestSourceRepository(db)
	mediaRepo := postgres.NewMediaRepository(db)
	sourceMediaRepo := postgres.NewSourceMediaRepository(db)
	dictRepo := postgres.NewDictionaryRepository(db)
	tokenRepo := postgres.NewAPITokenRepository(db)
	extRepo := postgres.NewMediaTypeExtensionRepository(db)

	minioCfg := &minio.Config{
		Host:     "localhost",
		Port:     9000,
		User:     "minioadmin",
		Password: "minioadmin",
		Bucket:   "media-parser",
		UseSSL:   false,
	}
	minioClient, err := minio.NewClient(minioCfg)
	s.Require().NoError(err)

	redisCfg := &config.RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "media_parser_password",
		DB:       0,
		CacheTTL: 24,
	}
	redisCacheRepo := redisCache.NewCacheRepository(redisCfg)

	parserCfg := &config.ParserConfig{
		PageLoadTimeout: 60,
		CacheTTLHours:   24,
		ParserWorkers:   5,
	}
	parseWorker := queue.NewParseWorker(
		sourceRepo, patternRepo, mediaRepo, sourceMediaRepo,
		requestRepo, requestSourceRepo, redisCacheRepo, dictRepo,
		extRepo, minioClient,
		parserCfg,
	)
	_ = parseWorker

	var producer *queue.Producer

	authService := service.NewAuthService(tokenRepo)
	sourceService := service.NewSourceService(sourceRepo, patternRepo, dictRepo)
	patternService := service.NewPatternService(patternRepo, dictRepo)
	requestService := service.NewRequestService(requestRepo, requestSourceRepo, dictRepo)
	mediaService := service.NewMediaService(mediaRepo, sourceMediaRepo, dictRepo)
	downloadService := service.NewDownloadService(mediaRepo, sourceMediaRepo, minioClient, 60*time.Second)
	dictionaryService := service.NewDictionaryService(dictRepo)
	parseService := service.NewParseService(
		requestService, mediaService, sourceService,
		sourceRepo, patternRepo, mediaRepo, sourceMediaRepo, redisCacheRepo,
		producer,
	)

	parseHandler := handler.NewParseHandler(parseService)
	mediaHandler := handler.NewMediaHandler(mediaService)
	requestHandler := handler.NewRequestHandler(requestService)
	sourceHandler := handler.NewSourceHandler(sourceService, patternService, parseService)
	downloadHandler := handler.NewDownloadHandler(downloadService)
	dictionaryHandler := handler.NewDictionaryHandler(dictionaryService)
	authMiddleware := middleware.NewAuthMiddleware(authService)

	gin.SetMode(gin.TestMode)
	s.router = gin.New()
	s.setupRoutes(
		parseHandler, mediaHandler, requestHandler, sourceHandler,
		downloadHandler, dictionaryHandler,
		authMiddleware,
	)
}

func (s *APITestSuite) TearDownSuite() {
	s.cleanupTestData()
	s.cancel()
	if s.dbPool != nil {
		s.dbPool.Close()
	}
	if s.redis != nil {
		s.redis.Close()
	}
}

func (s *APITestSuite) createTestToken() {
	query := `INSERT INTO api_tokens (token, name, active, created_at) VALUES ($1, $2, true, NOW())`
	_, err := s.dbPool.Exec(s.ctx, query, s.testToken, s.testID)
	s.Require().NoError(err)
}

func (s *APITestSuite) cleanupTestData() {
	queries := []string{
		`DELETE FROM source_media WHERE source_id IN (SELECT id FROM sources WHERE name LIKE 'test_%')`,
		`DELETE FROM request_sources WHERE source_id IN (SELECT id FROM sources WHERE name LIKE 'test_%')`,
		`DELETE FROM requests WHERE EXISTS (SELECT 1 FROM request_sources rs WHERE rs.request_id = requests.id AND rs.source_id IN (SELECT id FROM sources WHERE name LIKE 'test_%'))`,
		`DELETE FROM sources WHERE name LIKE 'test_%'`,
		`DELETE FROM patterns WHERE name LIKE 'test_%'`,
		`DELETE FROM media WHERE url LIKE '%test_%'`,
		`DELETE FROM api_tokens WHERE token = $1`,
	}

	for _, query := range queries {
		if query == `DELETE FROM api_tokens WHERE token = $1` {
			_, _ = s.dbPool.Exec(s.ctx, query, s.testToken)
		} else {
			_, _ = s.dbPool.Exec(s.ctx, query)
		}
	}

	iter := s.redis.Scan(s.ctx, 0, "*test_*", 100).Iterator()
	for iter.Next(s.ctx) {
		_ = s.redis.Del(s.ctx, iter.Val())
	}
}

func (s *APITestSuite) setupRoutes(
	parseHandler *handler.ParseHandler,
	mediaHandler *handler.MediaHandler,
	requestHandler *handler.RequestHandler,
	sourceHandler *handler.SourceHandler,
	downloadHandler *handler.DownloadHandler,
	dictionaryHandler *handler.DictionaryHandler,
	authMiddleware *middleware.AuthMiddleware,
) {
	v1 := s.router.Group("/api/v1")
	v1.Use(authMiddleware.Authenticate())
	{
		v1.POST("/parse/url", parseHandler.ParseURL)
		v1.POST("/parse/batch", parseHandler.ParseBatch)
		v1.POST("/parse/all", parseHandler.ParseAll)
		v1.POST("/parse/first", parseHandler.ParseFirst)
		v1.POST("/parse/n", parseHandler.ParseN)
		v1.POST("/parse/source", parseHandler.ParseSource)

		v1.GET("/media", mediaHandler.GetMediaList)
		v1.GET("/media/:id", mediaHandler.GetMediaByID)

		v1.GET("/requests", requestHandler.GetRequestList)
		v1.GET("/requests/:id", requestHandler.GetRequestByID)
		v1.GET("/requests/:id/media", requestHandler.GetRequestMedia)

		v1.GET("/sources", sourceHandler.GetSources)
		v1.POST("/sources", sourceHandler.CreateSource)
		v1.PUT("/sources/:id", sourceHandler.UpdateSource)
		v1.DELETE("/sources/:id", sourceHandler.DeleteSource)
		v1.POST("/sources/:id/parse", sourceHandler.ParseSource)

		v1.GET("/patterns", sourceHandler.GetPatterns)
		v1.POST("/patterns", sourceHandler.CreatePattern)
		v1.PUT("/patterns/:id", sourceHandler.UpdatePattern)
		v1.DELETE("/patterns/:id", sourceHandler.DeletePattern)

		v1.POST("/download/:id", downloadHandler.DownloadMedia)
		v1.POST("/download/url", downloadHandler.DownloadByURL)

		v1.GET("/dictionaries", dictionaryHandler.GetDictionaries)
		v1.GET("/dictionaries/media-types", dictionaryHandler.GetMediaTypes)
		v1.GET("/dictionaries/request-statuses", dictionaryHandler.GetRequestStatuses)
		v1.GET("/dictionaries/source-statuses", dictionaryHandler.GetSourceStatuses)
	}
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}

func (s *APITestSuite) performRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var w *httptest.ResponseRecorder
	var r *http.Request

	if body != nil {
		bodyBytes, _ := json.Marshal(body)
		r = httptest.NewRequest(method, path, bytes.NewReader(bodyBytes))
		r.Header.Set("Content-Type", "application/json")
	} else {
		r = httptest.NewRequest(method, path, nil)
	}

	r.Header.Set("X-Auth-Token", s.testToken)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, r)
	return w
}

// ============== DICTIONARIES TESTS ==============

func (s *APITestSuite) TestGetDictionaries() {
	w := s.performRequest(http.MethodGet, "/api/v1/dictionaries", nil)
	s.Equal(http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	s.True(resp["success"].(bool))
	data := resp["data"].(map[string]interface{})
	s.NotNil(data["media_types"])
	s.NotNil(data["request_statuses"])
	s.NotNil(data["source_statuses"])
}

func (s *APITestSuite) TestGetMediaTypes() {
	w := s.performRequest(http.MethodGet, "/api/v1/dictionaries/media-types", nil)
	s.Equal(http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	s.True(resp["success"].(bool))
	data := resp["data"].([]interface{})
	s.Greater(len(data), 0)
}

func (s *APITestSuite) TestGetRequestStatuses() {
	w := s.performRequest(http.MethodGet, "/api/v1/dictionaries/request-statuses", nil)
	s.Equal(http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	s.True(resp["success"].(bool))
	data := resp["data"].([]interface{})
	s.Greater(len(data), 0)
}

func (s *APITestSuite) TestGetSourceStatuses() {
	w := s.performRequest(http.MethodGet, "/api/v1/dictionaries/source-statuses", nil)
	s.Equal(http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	s.True(resp["success"].(bool))
	data := resp["data"].([]interface{})
	s.Greater(len(data), 0)
}

// ============== SOURCE TESTS ==============

func (s *APITestSuite) TestCreateSource() {
	req := map[string]interface{}{
		"name":     s.testID + "_source",
		"base_url": "https://test-example.com",
	}

	w := s.performRequest(http.MethodPost, "/api/v1/sources", req)
	s.Equal(http.StatusCreated, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	s.True(resp["success"].(bool))
}

func (s *APITestSuite) TestGetSources() {
	sourceReq := map[string]interface{}{
		"name":     s.testID + "_source_list",
		"base_url": "https://test-list.com",
	}
	s.performRequest(http.MethodPost, "/api/v1/sources", sourceReq)

	w := s.performRequest(http.MethodGet, "/api/v1/sources", nil)
	s.Equal(http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	s.True(resp["success"].(bool))
}

func (s *APITestSuite) TestUpdateSource() {
	createReq := map[string]interface{}{
		"name":     s.testID + "_source_update",
		"base_url": "https://test-update.com",
	}
	createResp := s.performRequest(http.MethodPost, "/api/v1/sources", createReq)
	var created map[string]interface{}
	json.Unmarshal(createResp.Body.Bytes(), &created)
	sourceID := int(created["data"].(map[string]interface{})["id"].(float64))

	updateReq := map[string]interface{}{
		"name":     s.testID + "_source_updated",
		"base_url": "https://test-updated.com",
	}
	w := s.performRequest(http.MethodPut, fmt.Sprintf("/api/v1/sources/%d", sourceID), updateReq)
	s.Equal(http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	s.True(resp["success"].(bool))
}

// ============== PATTERN TESTS ==============

func (s *APITestSuite) TestCreatePattern() {
	req := map[string]interface{}{
		"name":          s.testID + "_pattern",
		"regex":         `src="([^"]+\.jpg)"`,
		"media_type_id": 1,
	}

	w := s.performRequest(http.MethodPost, "/api/v1/patterns", req)
	s.Equal(http.StatusCreated, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	s.True(resp["success"].(bool))
}

// ============== PARSE TESTS ==============

func (s *APITestSuite) TestParseURL() {
	req := map[string]interface{}{
		"url":      "https://test-parse.com",
		"limit":    10,
		"offset":   0,
		"priority": 5,
	}

	w := s.performRequest(http.MethodPost, "/api/v1/parse/url", req)
	s.Equal(http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	s.True(resp["success"].(bool))
	data := resp["data"].(map[string]interface{})
	s.Equal("pending", data["status"])
	s.NotEmpty(data["request_id"])
}

func (s *APITestSuite) TestParseURLWithMediaType() {
	req := map[string]interface{}{
		"url":          "https://test-parse-media.com",
		"limit":        20,
		"offset":       10,
		"priority":     8,
		"media_type":   "image",
		"media_type_ids": []int{1, 2},
	}

	w := s.performRequest(http.MethodPost, "/api/v1/parse/url", req)
	s.Equal(http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	s.True(resp["success"].(bool))
}

func (s *APITestSuite) TestParseBatch() {
	req := map[string]interface{}{
		"urls":       []string{"https://test1.com", "https://test2.com"},
		"limit":      15,
		"offset":     5,
		"priority":   7,
		"media_type": "video",
	}

	w := s.performRequest(http.MethodPost, "/api/v1/parse/batch", req)
	s.Equal(http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	s.True(resp["success"].(bool))
	data := resp["data"].(map[string]interface{})
	s.Equal("pending", data["status"])
}

func (s *APITestSuite) TestParseAll() {
	req := map[string]interface{}{
		"limit":    25,
		"offset":   0,
		"priority": 3,
	}

	w := s.performRequest(http.MethodPost, "/api/v1/parse/all", req)
	s.Equal(http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	s.True(resp["success"].(bool))
}

func (s *APITestSuite) TestParseFirst() {
	req := map[string]interface{}{
		"priority": 9,
	}

	w := s.performRequest(http.MethodPost, "/api/v1/parse/first", req)
	s.Equal(http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	s.True(resp["success"].(bool))
}

func (s *APITestSuite) TestParseN() {
	req := map[string]interface{}{
		"count":    50,
		"offset":   10,
		"priority": 6,
	}

	w := s.performRequest(http.MethodPost, "/api/v1/parse/n", req)
	s.Equal(http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	s.True(resp["success"].(bool))
}

func (s *APITestSuite) TestParseSource() {
	sourceReq := map[string]interface{}{
		"name":     s.testID + "_source_parse",
		"base_url": "https://test-source-parse.com",
	}
	createResp := s.performRequest(http.MethodPost, "/api/v1/sources", sourceReq)
	var created map[string]interface{}
	json.Unmarshal(createResp.Body.Bytes(), &created)
	sourceID := int(created["data"].(map[string]interface{})["id"].(float64))

	parseReq := map[string]interface{}{
		"source_id":  sourceID,
		"limit":      30,
		"offset":     0,
		"priority":   4,
		"media_type": "image",
	}

	w := s.performRequest(http.MethodPost, "/api/v1/parse/source", parseReq)
	s.Equal(http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	s.True(resp["success"].(bool))
}

// ============== REQUEST TESTS ==============

func (s *APITestSuite) TestGetRequestByID() {
	parseReq := map[string]interface{}{
		"url":      "https://test-request.com",
		"limit":    5,
		"priority": 2,
	}
	parseResp := s.performRequest(http.MethodPost, "/api/v1/parse/url", parseReq)
	var parseResult map[string]interface{}
	json.Unmarshal(parseResp.Body.Bytes(), &parseResult)
	requestID := parseResult["data"].(map[string]interface{})["request_id"].(string)

	w := s.performRequest(http.MethodGet, fmt.Sprintf("/api/v1/requests/%s", requestID), nil)
	s.Equal(http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	s.True(resp["success"].(bool))
}

func (s *APITestSuite) TestGetRequestList() {
	w := s.performRequest(http.MethodGet, "/api/v1/requests?limit=10&offset=0", nil)
	s.Equal(http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	s.True(resp["success"].(bool))
}

// ============== DOWNLOAD TESTS ==============

func (s *APITestSuite) TestDownloadByURL() {
	req := map[string]interface{}{
		"url": "https://example.com/test.jpg",
	}

	w := s.performRequest(http.MethodPost, "/api/v1/download/url", req)
	s.NotEqual(http.StatusOK, w.Code)
}

// ============== NEGATIVE TESTS ==============

func (s *APITestSuite) TestUnauthorized() {
	r := httptest.NewRequest(http.MethodGet, "/api/v1/sources", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)
	s.Equal(http.StatusUnauthorized, w.Code)
}

func (s *APITestSuite) TestParseURLInvalidLimit() {
	req := map[string]interface{}{
		"url":   "https://test-invalid.com",
		"limit": 0,
	}

	w := s.performRequest(http.MethodPost, "/api/v1/parse/url", req)
	s.Equal(http.StatusBadRequest, w.Code)
}

func (s *APITestSuite) TestParseURLInvalidURL() {
	req := map[string]interface{}{
		"url":   "not-a-valid-url",
		"limit": 10,
	}

	w := s.performRequest(http.MethodPost, "/api/v1/parse/url", req)
	s.Equal(http.StatusBadRequest, w.Code)
}

func (s *APITestSuite) TestParseBatchEmptyURLs() {
	req := map[string]interface{}{
		"urls":  []string{},
		"limit": 10,
	}

	w := s.performRequest(http.MethodPost, "/api/v1/parse/batch", req)
	s.Equal(http.StatusBadRequest, w.Code)
}

func (s *APITestSuite) TestParseURLTooManyURLs() {
	urls := make([]string, 101)
	for i := 0; i < 101; i++ {
		urls[i] = fmt.Sprintf("https://test%d.com", i)
	}
	req := map[string]interface{}{
		"urls":  urls,
		"limit": 10,
	}

	w := s.performRequest(http.MethodPost, "/api/v1/parse/batch", req)
	s.Equal(http.StatusBadRequest, w.Code)
}

func (s *APITestSuite) TestHealth() {
	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, r)
	s.Equal(http.StatusOK, w.Code)
}

func TestAPI(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}
