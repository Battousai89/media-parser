package queue

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/media-parser/backend/internal/config"
	"github.com/media-parser/backend/internal/downloader"
	"github.com/media-parser/backend/internal/parser"
	"github.com/media-parser/backend/internal/repository"
	"github.com/media-parser/backend/internal/repository/minio"
)

type WorkerPools struct {
	parserPool   *ParserPool
	httpPool     *HTTPPool
	downloadPool *DownloadPool
	consumer     *Consumer
	wg           sync.WaitGroup
	ctx          context.Context
	cancel       context.CancelFunc
}

func NewWorkerPools(
	cfg *config.ParserConfig,
	consumer *Consumer,
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
	rabbitCfg *config.RabbitMQConfig,
) *WorkerPools {
	ctx, cancel := context.WithCancel(context.Background())

	parserPool := NewParserPool(
		cfg.ParserWorkers,
		sourceRepo,
		patternRepo,
		mediaRepo,
		sourceMediaRepo,
		requestRepo,
		requestSourceRepo,
		cacheRepo,
		dictRepo,
		extRepo,
		minioClient,
		cfg,
	)

	httpPool := NewHTTPPool(cfg.HTTPWorkers)

	downloadPool := NewDownloadPool(cfg.DownloadWorkers)

	return &WorkerPools{
		parserPool:   parserPool,
		httpPool:     httpPool,
		downloadPool: downloadPool,
		consumer:     consumer,
		ctx:          ctx,
		cancel:       cancel,
	}
}

func (wp *WorkerPools) Start() error {
	wp.parserPool.Start(wp.ctx)
	wp.httpPool.Start(wp.ctx)
	wp.downloadPool.Start(wp.ctx)

	wp.wg.Add(1)
	go wp.run()

	if err := wp.consumer.Start(); err != nil {
		return err
	}

	log.Printf("WorkerPools started: parser=%d, http=%d, download=%d",
		wp.parserPool.workers, wp.httpPool.workers, wp.downloadPool.workers)
	return nil
}

func (wp *WorkerPools) run() {
	defer wp.wg.Done()

	for {
		select {
		case task, ok := <-wp.consumer.ParseTasks():
			if !ok {
				return
			}
			wp.parserPool.Submit(task)

		case <-wp.ctx.Done():
			return
		}
	}
}

func (wp *WorkerPools) Stop() {
	log.Println("WorkerPools: shutting down...")

	wp.cancel()

	wp.parserPool.Stop()
	wp.httpPool.Stop()
	wp.downloadPool.Stop()
	wp.consumer.Stop()

	done := make(chan struct{})
	go func() {
		wp.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("WorkerPools: shutdown complete")
	case <-time.After(30 * time.Second):
		log.Println("WorkerPools: shutdown timeout")
	}
}

type ParserPool struct {
	workers           int
	taskChan          chan *ParseTask
	handler           *ParseWorker
	requestSourceRepo repository.RequestSourceRepository
	ctx               context.Context
	wg                sync.WaitGroup
	mu                sync.RWMutex
	isRunning         bool
}

func NewParserPool(
	workers int,
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
) *ParserPool {
	return &ParserPool{
		workers:           workers,
		taskChan:          make(chan *ParseTask, workers*2),
		handler: NewParseWorker(
			sourceRepo,
			patternRepo,
			mediaRepo,
			sourceMediaRepo,
			requestRepo,
			requestSourceRepo,
			cacheRepo,
			dictRepo,
			extRepo,
			minioClient,
			cfg,
		),
		requestSourceRepo: requestSourceRepo,
	}
}

func (p *ParserPool) Start(ctx context.Context) {
	p.mu.Lock()
	if p.isRunning {
		p.mu.Unlock()
		return
	}
	p.isRunning = true
	p.ctx = ctx
	p.mu.Unlock()

	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}

	log.Printf("ParserPool: started %d workers", p.workers)
}

func (p *ParserPool) worker(id int) {
	defer p.wg.Done()

	for {
		select {
		case task, ok := <-p.taskChan:
			if !ok {
				return
			}
			if err := p.handler.Handle(p.ctx, task); err != nil {
				log.Printf("ParserPool worker %d: task %s error: %v (retry=%d/%d)", id, task.RequestID, err, task.RetryCount, task.MaxRetries)
				if task.RetryCount < task.MaxRetries {
					task.RetryCount++
					// Сохраняем retry_count в БД
					_ = p.requestSourceRepo.IncrementRetryCount(p.ctx, task.RequestID, task.SourceID)
					log.Printf("ParserPool worker %d: retrying task %s (retry=%d)", id, task.RequestID, task.RetryCount)
					time.Sleep(5 * time.Second)
					p.Submit(task)
				} else {
					// Max retries reached - check if request should be marked as failed/partial
					log.Printf("ParserPool worker %d: max retries reached for task %s, calling checkRequestCompletion", id, task.RequestID)
					p.handler.checkRequestCompletion(p.ctx, task.RequestID)
				}
			}
		case <-p.ctx.Done():
			return
		}
	}
}

func (p *ParserPool) Submit(task *ParseTask) {
	p.mu.RLock()
	running := p.isRunning
	p.mu.RUnlock()

	if !running {
		return
	}

	select {
	case p.taskChan <- task:
	default:
		log.Printf("ParserPool: task channel full, dropping task %s", task.RequestID)
	}
}

func (p *ParserPool) Stop() {
	p.mu.Lock()
	if !p.isRunning {
		p.mu.Unlock()
		return
	}
	p.isRunning = false
	p.mu.Unlock()

	close(p.taskChan)

	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(30 * time.Second):
	}
}

type HTTPPool struct {
	workers   int
	taskChan  chan *HTTPFetchTask
	client    *parser.HTTPClient
	ctx       context.Context
	wg        sync.WaitGroup
	mu        sync.RWMutex
	isRunning bool
}

func NewHTTPPool(workers int) *HTTPPool {
	return &HTTPPool{
		workers:  workers,
		taskChan: make(chan *HTTPFetchTask, workers*2),
		client:   parser.NewHTTPClient(parser.HTTPClientConfig{Timeout: 30 * time.Second}),
	}
}

func (h *HTTPPool) Start(ctx context.Context) {
	h.mu.Lock()
	if h.isRunning {
		h.mu.Unlock()
		return
	}
	h.isRunning = true
	h.ctx = ctx
	h.mu.Unlock()

	for i := 0; i < h.workers; i++ {
		h.wg.Add(1)
		go h.worker(i)
	}

	log.Printf("HTTPPool: started %d workers", h.workers)
}

func (h *HTTPPool) worker(id int) {
	defer h.wg.Done()

	for {
		select {
		case task, ok := <-h.taskChan:
			if !ok {
				return
			}
			_, _ = h.client.Fetch(h.ctx, task.URL)
		case <-h.ctx.Done():
			return
		}
	}
}

func (h *HTTPPool) Submit(task *HTTPFetchTask) {
	h.mu.RLock()
	running := h.isRunning
	h.mu.RUnlock()

	if !running {
		return
	}

	select {
	case h.taskChan <- task:
	default:
	}
}

func (h *HTTPPool) Stop() {
	h.mu.Lock()
	if !h.isRunning {
		h.mu.Unlock()
		return
	}
	h.isRunning = false
	h.mu.Unlock()

	close(h.taskChan)

	done := make(chan struct{})
	go func() {
		h.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(30 * time.Second):
	}
}

type DownloadPool struct {
	workers    int
	taskChan   chan *DownloadTask
	downloader *downloader.Downloader
	ctx        context.Context
	wg         sync.WaitGroup
	mu        sync.RWMutex
	isRunning  bool
}

func NewDownloadPool(workers int) *DownloadPool {
	return &DownloadPool{
		workers:    workers,
		taskChan:   make(chan *DownloadTask, workers*2),
		downloader: downloader.NewDownloader(downloader.Config{Timeout: 60 * time.Second}),
	}
}

func (d *DownloadPool) Start(ctx context.Context) {
	d.mu.Lock()
	if d.isRunning {
		d.mu.Unlock()
		return
	}
	d.isRunning = true
	d.ctx = ctx
	d.mu.Unlock()

	for i := 0; i < d.workers; i++ {
		d.wg.Add(1)
		go d.worker(i)
	}

	log.Printf("DownloadPool: started %d workers", d.workers)
}

func (d *DownloadPool) worker(id int) {
	defer d.wg.Done()

	for {
		select {
		case task, ok := <-d.taskChan:
			if !ok {
				return
			}
			_, _ = d.downloader.Download(d.ctx, task.MediaURL)
		case <-d.ctx.Done():
			return
		}
	}
}

func (d *DownloadPool) Submit(task *DownloadTask) {
	d.mu.RLock()
	running := d.isRunning
	d.mu.RUnlock()

	if !running {
		return
	}

	select {
	case d.taskChan <- task:
	default:
	}
}

func (d *DownloadPool) Stop() {
	d.mu.Lock()
	if !d.isRunning {
		d.mu.Unlock()
		return
	}
	d.isRunning = false
	d.mu.Unlock()

	close(d.taskChan)

	done := make(chan struct{})
	go func() {
		d.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(30 * time.Second):
	}
}
