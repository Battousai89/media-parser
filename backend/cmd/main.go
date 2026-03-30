package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/media-parser/backend/docs/swagger"
	"github.com/media-parser/backend/internal/config"
	"github.com/media-parser/backend/internal/handler"
	"github.com/media-parser/backend/internal/handler/middleware"
	"github.com/media-parser/backend/internal/queue"
	"github.com/media-parser/backend/internal/repository/minio"
	"github.com/media-parser/backend/internal/repository/postgres"
	redisCache "github.com/media-parser/backend/internal/repository/redis"
	"github.com/media-parser/backend/internal/service"
)

// @title Media Parser API
// @version 1.0
// @description Backend for parsing media from the internet
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-Auth-Token

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := postgres.NewPostgresDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	redisRepo := redisCache.NewCacheRepository(&cfg.Redis)
	if err := redisRepo.Ping(context.Background()); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisRepo.Close()

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
		Host:     cfg.Minio.Host,
		Port:     cfg.Minio.Port,
		User:     cfg.Minio.User,
		Password: cfg.Minio.Password,
		Bucket:   cfg.Minio.Bucket,
		UseSSL:   cfg.Minio.UseSSL,
	}
	minioClient, err := minio.NewClient(minioCfg)
	if err != nil {
		log.Fatalf("Failed to connect to Minio: %v", err)
	}

	authService := service.NewAuthService(tokenRepo)
	sourceService := service.NewSourceService(sourceRepo, patternRepo, dictRepo)
	patternService := service.NewPatternService(patternRepo, dictRepo)
	requestService := service.NewRequestService(requestRepo, requestSourceRepo, dictRepo)
	mediaService := service.NewMediaService(mediaRepo, sourceMediaRepo, dictRepo)
	downloadService := service.NewDownloadService(mediaRepo, sourceMediaRepo, minioClient, 60*time.Second)
	dictionaryService := service.NewDictionaryService(dictRepo)

	producer := queue.NewProducer(&cfg.RabbitMQ)
	if err := producer.Connect(); err != nil {
		log.Printf("Warning: Failed to connect to RabbitMQ producer: %v", err)
	}
	defer producer.Close()

	parseService := service.NewParseService(
		requestService,
		mediaService,
		sourceService,
		sourceRepo,
		patternRepo,
		mediaRepo,
		sourceMediaRepo,
		redisRepo,
		producer,
	)

	parseHandler := handler.NewParseHandler(parseService)
	mediaHandler := handler.NewMediaHandler(mediaService)
	requestHandler := handler.NewRequestHandler(requestService)
	sourceHandler := handler.NewSourceHandler(sourceService, patternService, parseService)
	downloadHandler := handler.NewDownloadHandler(downloadService)
	dictionaryHandler := handler.NewDictionaryHandler(dictionaryService)

	authMiddleware := middleware.NewAuthMiddleware(authService)
	rateLimitMiddleware := middleware.NewPerKeyRateLimitMiddleware(middleware.RateLimitConfig{
		RequestsPerSecond: 10,
		BurstSize:         20,
	})
	loggerMiddleware := middleware.NewLoggerMiddleware()

	gin.SetMode(cfg.Server.Mode)
	router := gin.New()
	router.Use(middleware.Recovery())
	router.Use(loggerMiddleware.Logger())
	router.Use(middleware.CORS())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
		ginSwagger.DefaultModelsExpandDepth(-1),
		ginSwagger.InstanceName(swagger.SwaggerInfo.InstanceName())))

	v1 := router.Group("/api/v1")
	v1.Use(authMiddleware.Authenticate())
	v1.Use(rateLimitMiddleware.Limit())
	{
		parse := v1.Group("/parse")
		{
			parse.POST("/url", parseHandler.ParseURL)
			parse.POST("/batch", parseHandler.ParseBatch)
			parse.POST("/all", parseHandler.ParseAll)
			parse.POST("/first", parseHandler.ParseFirst)
			parse.POST("/n", parseHandler.ParseN)
		}

		media := v1.Group("/media")
		{
			media.GET("", mediaHandler.GetMediaList)
			media.GET("/:id", mediaHandler.GetMediaByID)
			media.POST("/upload", mediaHandler.UploadMedia)
			media.POST("/check", mediaHandler.CheckMedia)
			media.DELETE("/:id", mediaHandler.DeleteMedia)
		}

		download := v1.Group("/download")
		{
			download.POST("/:id", downloadHandler.DownloadMedia)
			download.POST("/url", downloadHandler.DownloadByURL)
		}

		requests := v1.Group("/requests")
		{
			requests.GET("", requestHandler.GetRequestList)
			requests.GET("/:id", requestHandler.GetRequestByID)
			requests.GET("/:id/media", requestHandler.GetRequestMedia)
			requests.DELETE("/:id", requestHandler.DeleteRequest)
		}

		sources := v1.Group("/sources")
		{
			sources.GET("", sourceHandler.GetSources)
			sources.POST("", sourceHandler.CreateSource)
			sources.PUT("/:id", sourceHandler.UpdateSource)
			sources.DELETE("/:id", sourceHandler.DeleteSource)
			sources.POST("/:id/parse", sourceHandler.ParseSource)
		}

		patterns := v1.Group("/patterns")
		{
			patterns.GET("", sourceHandler.GetPatterns)
			patterns.POST("", sourceHandler.CreatePattern)
			patterns.PUT("/:id", sourceHandler.UpdatePattern)
			patterns.DELETE("/:id", sourceHandler.DeletePattern)
		}

		dictionaries := v1.Group("/dictionaries")
		{
			dictionaries.GET("", dictionaryHandler.GetDictionaries)
			dictionaries.GET("/media-types", dictionaryHandler.GetMediaTypes)
			dictionaries.GET("/request-statuses", dictionaryHandler.GetRequestStatuses)
			dictionaries.GET("/source-statuses", dictionaryHandler.GetSourceStatuses)
		}
	}

	consumer := queue.NewConsumer(&cfg.RabbitMQ)
	if err := consumer.Connect(); err != nil {
		log.Printf("Warning: Failed to connect to RabbitMQ consumer: %v", err)
	}

	workerPools := queue.NewWorkerPools(
		&cfg.Parser,
		consumer,
		sourceRepo,
		patternRepo,
		mediaRepo,
		sourceMediaRepo,
		requestRepo,
		requestSourceRepo,
		redisRepo,
		dictRepo,
		extRepo,
		minioClient,
		&cfg.RabbitMQ,
	)

	if err := workerPools.Start(); err != nil {
		log.Printf("Warning: Failed to start worker pools: %v", err)
	} else {
		log.Println("Worker pools started")
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Starting server on port %d...", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	workerPools.Stop()

	log.Println("Server stopped")
}
