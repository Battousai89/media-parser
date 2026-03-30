# План исправлений медиа-парсера

## Обзор

Документ описывает критические проблемы в текущей реализации и план их исправления.

---

## 1. Лимит запроса: хардкод вместо параметров

### Проблема
```go
req, err := s.requestService.CreateRequest(ctx, mediaTypeCode, limit*len(urls), 0, 0, tokenID)
//                                                   ^^^^^^^^^^^^  ^  ^
//                                                   limit умножается на кол-во URL,
//                                                   offset и priority = 0 (хардкод)
```

**Почему неверно:**
- `offset` всегда 0 — нельзя запросить "вторую страницу" результатов
- `priority` всегда 0 — нельзя задать приоритет запроса
- `limit*len(urls)` — неправильная логика: если запрошено 10 медиа с 2 сайтов, должно быть 10, а не 20

### Решение

**1.1. Обновить DTO**
```go
// internal/model/dto/parse.go
type ParseBatchRequest struct {
    URLs      []string `json:"urls" binding:"required,min=1,max=100"`
    MediaType *string  `json:"media_type,omitempty"`
    Limit     int      `json:"limit" binding:"required,min=1,max=1000"`  // не *int
    Offset    int      `json:"offset" binding:"omitempty,min=0"`         // новое поле
    Priority  int      `json:"priority" binding:"omitempty,min=0,max=10"`// новое поле
    Download  *bool    `json:"download,omitempty"`
}
```

**1.2. Обновить сервис**
```go
// internal/service/parse_service.go
func (s *ParseService) ParseBatch(
    ctx context.Context,
    urls []string,
    mediaTypeCode *string,
    limit int,       // не умножать
    offset int,      // новый параметр
    priority int,    // новый параметр
    download bool,
    tokenID *int,
) (*entity.Request, error) {
    req, err := s.requestService.CreateRequest(
        ctx, mediaTypeCode, limit, offset, priority, tokenID)
    // ...
}
```

**1.3. Обновить handler**
```go
// internal/handler/parse_handler.go
limit := req.Limit
if limit == 0 {
    limit = 10  // дефолт только если не указан
}
offset := req.Offset
priority := req.Priority

result, err := h.parseService.ParseBatch(
    c, req.URLs, req.MediaType, limit, offset, priority, download, tokenID)
```

---

## 2. MediaTypeID в запросе: поддержка нескольких типов

### Проблема
```go
MediaTypeID: mediaType.ID  // одно значение, а нужно несколько
```

**Почему неверно:**
- Пользователь может хотеть парсить "image И video"
- Таблица `requests.media_type_id` позволяет только один тип

### Решение

**2.1. Миграция: таблица request_media_types**
```sql
-- migrations/011_request_media_types.sql
-- +goose Up

CREATE TABLE request_media_types (
    id            SERIAL PRIMARY KEY,
    request_id    UUID NOT NULL REFERENCES requests(id) ON DELETE CASCADE,
    media_type_id INT NOT NULL REFERENCES media_types(id),
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(request_id, media_type_id)
);

CREATE INDEX idx_request_media_types_request ON request_media_types(request_id);

-- Перенос данных из requests.media_type_id
INSERT INTO request_media_types (request_id, media_type_id)
SELECT id, media_type_id FROM requests WHERE media_type_id IS NOT NULL;

-- +goose Down
DROP TABLE IF EXISTS request_media_types;
```

**2.2. Обновить entity.Request**
```go
// internal/model/entity/request.go
type Request struct {
    ID              uuid.UUID      `db:"id" json:"id"`
    StatusID        int            `db:"status_id" json:"status_id"`
    MediaTypeIDs    []int          `db:"-" json:"media_type_ids,omitempty"` // новое поле
    MediaTypeID     *int           `db:"media_type_id" json:"-"`            // устаревает
    // ...
}
```

**2.3. Обновить repository**
```go
// internal/repository/postgres/request_repo.go
type RequestRepository interface {
    // ...
    GetMediaTypeIDs(ctx context.Context, requestID uuid.UUID) ([]int, error)
    SetMediaTypeIDs(ctx context.Context, requestID uuid.UUID, mediaTypeIDs []int) error
}
```

**2.4. Обновить сервис**
```go
// internal/service/request_service.go
func (s *RequestService) CreateRequest(
    ctx context.Context,
    mediaTypeCodes []string,  // не один код, а список
    limit, offset, priority int,
    tokenID *int,
) (*entity.Request, error) {
    // ...
    // Для каждого кода получить ID и сохранить в request_media_types
}
```

---

## 3. Статус источника: хардкод "active" вместо проверки

### Проблема
```go
source := &entity.Source{
    Name:     name,
    BaseURL:  baseURL,
    StatusID: 1,  // хардкод "active"
}
```

**Почему неверно:**
- Источник может быть недоступен (404, 503, timeout)
- Источник может быть заблокирован (robots.txt, DNS block)
- Статус должен определяться реальной проверкой

### Решение

**3.1. Обновить SourceService.GetOrCreateSource**
```go
// internal/service/source_service.go
func (s *SourceService) GetOrCreateSource(ctx context.Context, url string) (*entity.Source, error) {
    source, err := s.sourceRepo.GetByURL(ctx, url)
    if err != nil {
        return nil, err
    }
    if source != nil {
        // Проверить актуальность статуса
        status, err := s.checkSourceStatus(ctx, url)
        if err != nil {
            status = entity.SourceStatusError
        }
        if status != source.StatusID {
            source.StatusID = status
            _ = s.sourceRepo.Update(ctx, source)
        }
        return source, nil
    }

    // Новый источник: проверить статус перед созданием
    statusCode, err := s.checkSourceStatus(ctx, url)
    if err != nil {
        statusCode = entity.SourceStatusError  // 3
    }

    name := url
    if len(name) > 255 {
        name = name[:255]
    }
    return s.Create(ctx, name, url, &statusCode)
}

func (s *SourceService) checkSourceStatus(ctx context.Context, url string) (int, error) {
    // 1. Проверка DNS
    // 2. HEAD-запрос с таймаутом 5s
    // 3. Анализ статуса:
    //    - 200-299 → active (1)
    //    - 403, 429 → blocked (4)
    //    - 404, 500-599 → error (3)
    //    - timeout → error (3)
}
```

**3.2. Добавить миграцию (если нужна)**
```sql
-- migrations/012_add_source_check_fields.sql
-- +goose Up

ALTER TABLE sources ADD COLUMN last_checked_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE sources ADD COLUMN check_error_message TEXT;

-- +goose Down
ALTER TABLE sources DROP COLUMN IF EXISTS last_checked_at;
ALTER TABLE sources DROP COLUMN IF EXISTS check_error_message;
```

---

## 4. Параметр Download: бессмысленный флаг

### Проблема
```go
Download  *bool  `json:"download,omitempty"`
```

**Почему неверно:**
- Мы храним только URL медиа, не скачиваем файлы автоматически
- Флаг создаёт ложное ожидание у клиента
- В коде нет реальной логики скачивания при парсинге

### Решение

**4.1. Удалить параметр из DTO**
```go
// internal/model/dto/parse.go
type ParseURLRequest struct {
    URL       string  `json:"url" binding:"required,url"`
    MediaType *string `json:"media_type,omitempty"`
    Limit     int     `json:"limit" binding:"required,min=1,max=1000"`
    Offset    int     `json:"offset" binding:"omitempty,min=0"`
    Priority  int     `json:"priority" binding:"omitempty,min=0,max=10"`
    // Download удалён
}
```

**4.2. Удалить из сервиса**
```go
// internal/service/parse_service.go
func (s *ParseService) ParseURL(
    ctx context.Context,
    url string,
    mediaTypeCode *string,
    limit int,
    offset int,
    priority int,
    tokenID *int,
) (*entity.Request, error) {
    // ...
}
```

**4.3. Создать отдельный endpoint для скачивания**
```go
// internal/handler/download_handler.go
type DownloadHandler struct {
    downloadService *service.DownloadService
}

// @Summary Скачать медиа
// @Tags download
// @Produce application/octet-stream
// @Param id path string true "Media ID"
// @Success 200 {file} binary
// @Router /api/v1/download/:id [post]
func (h *DownloadHandler) DownloadMedia(c *gin.Context) {
    id, _ := uuid.Parse(c.Param("id"))
    media, _ := h.downloadService.GetMediaByID(c, id)
    
    result, err := h.downloadService.Download(c, media.URL)
    if err != nil {
        c.JSON(500, dto.Response{Success: false, Error: ...})
        return
    }
    
    c.FileAttachment(result.FilePath, "download")
}
```

**4.4. Создать DownloadService**
```go
// internal/service/download_service.go
type DownloadService struct {
    mediaRepo       repository.MediaRepository
    sourceMediaRepo repository.SourceMediaRepository
    downloader      *downloader.Downloader
}

func (s *DownloadService) Download(ctx context.Context, url string) (*DownloadResult, error) {
    return s.downloader.Download(ctx, url)
}
```

---

## 5. TTL кэша: хардкод 24 часа

### Проблема
```go
if cached != nil && time.Since(cached.ParsedAt) < 24*time.Hour {
    // хардкод
}

cacheRepo.SetURLCache(ctx, task.URL, hash, 24*time.Hour)
```

**Почему неверно:**
- Нельзя изменить TTL без пересборки
- В `.env` есть `CACHE_TTL_HOURS=24`, но оно не используется

### Решение

**5.1. Передать TTL в ParseWorker**
```go
// internal/queue/parse_worker.go
type ParseWorker struct {
    // ...
    cacheTTL time.Duration
}

func NewParseWorker(
    // ...
    cfg *config.ParserConfig,
) *ParseWorker {
    cacheTTL := time.Duration(cfg.CacheTTLHours) * time.Hour
    return &ParseWorker{
        // ...
        cacheTTL: cacheTTL,
    }
}
```

**5.2. Обновить config**
```go
// internal/config/config.go
type ParserConfig struct {
    // ...
    CacheTTLHours int `env:"CACHE_TTL_HOURS" envDefault:"24"`
}
```

**5.3. Использовать в worker**
```go
// internal/queue/parse_worker.go
func (w *ParseWorker) Handle(ctx context.Context, task *ParseTask) error {
    // ...
    
    cached, err := w.cacheRepo.GetURLCache(ctx, task.URL)
    if err == nil && cached != nil && time.Since(cached.ParsedAt) < w.cacheTTL {
        log.Printf("ParseWorker: URL %s is cached, skipping", task.URL)
        return nil
    }
    
    // ...
    
    _ = w.cacheRepo.SetURLCache(ctx, task.URL, w.hashURL(task.URL), w.cacheTTL)
}
```

---

## 6. Статусы источников и запросов: неверная логика

### Проблема

**6.1. Статус источника "completed"**
```go
statusCode := "completed"  // такого статуса нет в source_statuses!
requestSourceRepo.UpdateStatus(ctx, ..., statusCode, ...)
```

**Почему неверно:**
- `source_statuses`: active, inactive, error, blocked
- `request_statuses`: pending, processing, completed, failed, partial
- Это разные справочники!

**6.2. Логика обновления статуса**
```go
// Статус источника должен определяться при ПЕРВОЙ попытке доступа
// а не после каждого парсинга
```

### Решение

**6.1. Разделить статусы**
```go
// internal/model/entity/dictionary.go
const (
    // Статусы запросов
    RequestStatusPending    = "pending"
    RequestStatusProcessing = "processing"
    RequestStatusCompleted  = "completed"
    RequestStatusFailed     = "failed"
    RequestStatusPartial    = "partial"
    
    // Статусы источников
    SourceStatusActive   = "active"
    SourceStatusInactive = "inactive"
    SourceStatusError    = "error"
    SourceStatusBlocked  = "blocked"
)
```

**6.2. Обновить логику ParseWorker**
```go
// internal/queue/parse_worker.go
func (w *ParseWorker) Handle(ctx context.Context, task *ParseTask) error {
    // 1. Обновить статус запроса в processing
    _ = w.requestRepo.UpdateStatus(ctx, task.RequestID, w.statusCode("processing"), nil)
    
    // 2. Проверка кэша/robots.txt/HTTP — это проверка доступности источника
    canFetch, err := w.robotsChecker.CanFetch(task.URL)
    if err != nil || !canFetch {
        // Источник заблокирован
        _ = w.sourceRepo.UpdateStatus(ctx, task.SourceID, w.sourceStatusCode("blocked"))
        _ = w.requestSourceRepo.UpdateStatus(ctx, task.RequestID, task.SourceID, 
            w.statusCode("failed"), 0, &errMsg)
        return fmt.Errorf("blocked by robots.txt")
    }
    
    result, err := w.httpClient.Fetch(ctx, task.URL)
    if err != nil {
        // Источник недоступен
        _ = w.sourceRepo.UpdateStatus(ctx, task.SourceID, w.sourceStatusCode("error"))
        _ = w.requestSourceRepo.UpdateStatus(ctx, task.RequestID, task.SourceID,
            w.statusCode("failed"), 0, &errMsg)
        return err
    }
    
    // 3. Источник доступен — обновить статус на active
    _ = w.sourceRepo.UpdateStatus(ctx, task.SourceID, w.sourceStatusCode("active"))
    
    // 4. Парсинг и сохранение медиа
    parsedCount := w.parseAndSave(ctx, result.Body, task)
    
    // 5. Обновить request_sources: статус не менять, только parsed_count
    _ = w.requestSourceRepo.UpdateParsedCount(ctx, task.RequestID, task.SourceID, parsedCount)
    
    // 6. Проверить завершение всех источников запроса
    w.checkRequestCompletion(ctx, task.RequestID)
    
    return nil
}
```

**6.3. Обновить repository**
```go
// internal/repository/postgres/source_repo.go
type SourceRepository interface {
    // ...
    UpdateStatus(ctx context.Context, id int, statusID int) error
}

func (r *SourceRepositoryImpl) UpdateStatus(ctx context.Context, id int, statusID int) error {
    query := `UPDATE sources SET status_id = $1, updated_at = NOW() WHERE id = $2`
    _, err := r.db.Pool.Exec(ctx, query, statusID, id)
    return err
}
```

**6.4. Проверка завершения запроса**
```go
func (w *ParseWorker) checkRequestCompletion(ctx context.Context, requestID uuid.UUID) {
    sources, _ := w.requestSourceRepo.GetByRequestID(ctx, requestID)
    
    totalParsed := 0
    allProcessed := true
    
    for _, s := range sources {
        if s.StatusID == w.statusCode("processing") {
            allProcessed = false
            break
        }
        totalParsed += s.ParsedCount
    }
    
    if allProcessed {
        statusCode := "completed"
        if totalParsed == 0 {
            statusCode = "partial"
        }
        _ = w.requestRepo.UpdateStatus(ctx, requestID, w.statusCode(statusCode), nil)
    }
}
```

---

## 7. Приоритет, лимит, оффсет, retry_count: не работают

### Проблема
```go
// В БД есть поля:
priority     INT NOT NULL DEFAULT 0
retry_count  INT NOT NULL DEFAULT 0
max_retries  INT NOT NULL DEFAULT 3

// Но нигде не используются:
// - priority не влияет на порядок обработки
// - retry_count не инкрементится при ошибках
// - max_retries не проверяется
```

### Решение

**7.1. Приоритет в RabbitMQ**
```go
// internal/queue/producer.go
func (p *Producer) Publish(ctx context.Context, task *ParseTask) error {
    // ...
    
    msg := amqp091.Publishing{
        DeliveryMode: amqp091.Persistent,
        ContentType:  "application/json",
        Body:         body,
        Priority:     uint8(task.Priority),  // 0-9
        MessageId:    uuid.New().String(),
    }
    
    return p.channel.PublishWithContext(ctx, p.exchange, p.config.Queue, false, false, msg)
}
```

**7.2. Очередь с приоритетами**
```sql
-- migrations/013_rabbitmq_priority_queue.sql
-- Нужно пересоздать очередь с поддержкой приоритетов
-- (выполняется вручную в RabbitMQ)

-- declare queue с x-max-priority: 10
queue, err := channel.QueueDeclare(
    p.config.Queue,
    true,
    false,
    false,
    false,
    amqp091.Table{
        "x-max-priority": 10,
    },
)
```

**7.3. Retry count при ошибках**
```go
// internal/queue/parse_worker.go
func (w *ParseWorker) Handle(ctx context.Context, task *ParseTask) error {
    // ...
    
    if err != nil {
        // Инкремент retry_count
        req, _ := w.requestRepo.GetByID(ctx, task.RequestID)
        if req.RetryCount < req.MaxRetries {
            _ = w.requestRepo.IncrementRetryCount(ctx, task.RequestID)
            
            // Повтор через 5 секунд
            time.Sleep(5 * time.Second)
            return w.Handle(ctx, task)  // рекурсивный вызов
        }
        
        // Превышено кол-во попыток
        errMsg := fmt.Sprintf("max retries exceeded: %v", err)
        _ = w.requestRepo.UpdateStatus(ctx, task.RequestID, w.statusCode("failed"), &errMsg)
        return err
    }
}
```

**7.4. Repository методы**
```go
// internal/repository/postgres/request_repo.go
func (r *RequestRepositoryImpl) IncrementRetryCount(ctx context.Context, id uuid.UUID) error {
    query := `
        UPDATE requests 
        SET retry_count = retry_count + 1, updated_at = NOW()
        WHERE id = $1
    `
    _, err := r.db.Pool.Exec(ctx, query, id)
    return err
}
```

**7.5. Обработка по приоритету в Consumer**
```go
// internal/queue/consumer.go
// RabbitMQ сам обрабатывает приоритеты при delivery
// нужно только установить x-max-priority при создании очереди
```

---

## 8. Selenium: не используется вообще

### Проблема

**Код Selenium есть, но НЕ вызывается нигде:**
```bash
# Поиск вызовов NewDriver или GetPageSource
$ grep -r "NewDriver\|GetPageSource" --include="*.go" .
# Результат: только определения в chrome.go, вызовов нет
```

**В ParseWorker используется только HTTP-клиент:**
```go
// internal/queue/parse_worker.go
result, err := w.httpClient.Fetch(ctx, task.URL)  // ← обычный HTTP GET, без JS
```

**Почему это критично:**
- Сайты на React/Vue/Angular **не парсятся** — контент подгружается через JS
- Lazy-loading изображения **не видны** — подгружаются при скролле
- Бесконечные ленты (infinite scroll) **не работают**
- Контент за paywall/авторизацией **недоступен**

**Пример:**
```
https://pikabu.ru — пустой HTML без Selenium:
  <div id="app"></div>  ← контент рендерится в браузере

https://lenta.ru — работает через HTTP:
  <article>...</article>  ← статический HTML
```

### Решение

**8.1. Добавить SeleniumDriver в ParseWorker**
```go
// internal/queue/parse_worker.go
type ParseWorker struct {
    // ...
    httpClient     *parser.HTTPClient
    seleniumDriver *selenium.Driver
}

func NewParseWorker(
    // ...
) *ParseWorker {
    w := &ParseWorker{
        // ...
        httpClient: parser.NewHTTPClient(parser.HTTPClientConfig{
            Timeout: 30 * time.Second,
        }),
        seleniumDriver: selenium.NewDriver(selenium.Config{
            Host:            cfg.SeleniumHost,
            Port:            cfg.SeleniumPort,
            Headless:        cfg.SeleniumHeadless,
            PageLoadTimeout: cfg.PageLoadTimeout,
        }),
    }
    return w
}
```

**8.2. Обновить Handle() для использования Selenium**
```go
// internal/queue/parse_worker.go
func (w *ParseWorker) Handle(ctx context.Context, task *ParseTask) error {
    // ...
    
    // Всегда используем Selenium для JS-рендеринга
    log.Printf("ParseWorker: fetching URL %s via Selenium", task.URL)
    html, err := w.seleniumDriver.GetPageSource(ctx, task.URL)
    if err != nil {
        errMsg := err.Error()
        log.Printf("ParseWorker: Selenium error: %v", err)
        _ = w.requestSourceRepo.UpdateStatus(ctx, task.RequestID, task.SourceID, 
            w.statusCode("failed"), 0, &errMsg)
        return fmt.Errorf("fetch page via Selenium: %w", err)
    }
    
    log.Printf("ParseWorker: fetched %d bytes via Selenium", len(html))
    
    // ...
}
```

**8.3. Удалить неиспользуемый HTTP-клиент из ParseWorker**
```go
// Оставить HTTP-клиент только для проверки доступности (HEAD-запрос)
// Основной fetch — только через Selenium
```

**8.4. Обновить config**
```go
// internal/config/config.go
type ParserConfig struct {
    SeleniumHost         string `env:"SELENIUM_HOST" envDefault:"selenium"`
    SeleniumPort         int    `env:"SELENIUM_PORT" envDefault:"4444"`
    SeleniumHeadless     bool   `env:"SELENIUM_HEADLESS" envDefault:"true"`
    PageLoadTimeout      int    `env:"PAGE_LOAD_TIMEOUT" envDefault:"60"`
    // ...
}
```

**8.5. Обновить worker_pools.go**
```go
// internal/queue/worker_pools.go
func NewParserPool(
    workers int,
    // ...
    cfg *config.ParserConfig,  // ← добавить для конфига Selenium
) *ParserPool {
    return &ParserPool{
        workers:  workers,
        taskChan: make(chan *ParseTask, workers*2),
        handler: NewParseWorker(
            // ...
            cfg,  // ← передать конфиг
        ),
    }
}
```

**8.6. Производительность: пул Selenium-драйверов**

Проблема: `selenium.WebDriver` тяжёлый, нельзя создавать на каждый запрос.

Решение: пул переиспользуемых драйверов.

```go
// internal/parser/selenium/pool.go
type DriverPool struct {
    drivers    []*Driver
    idleChan   chan *Driver
    busyChan   chan *Driver
    mu         sync.Mutex
    maxDrivers int
}

func NewDriverPool(cfg Config, maxDrivers int) *DriverPool {
    pool := &DriverPool{
        drivers:    make([]*Driver, 0, maxDrivers),
        idleChan:   make(chan *Driver, maxDrivers),
        busyChan:   make(chan *Driver, maxDrivers),
        maxDrivers: maxDrivers,
    }
    
    // Предварительно создать N драйверов
    for i := 0; i < maxDrivers; i++ {
        d := NewDriver(cfg)
        _ = d.Init(context.Background())
        pool.drivers = append(pool.drivers, d)
        pool.idleChan <- d
    }
    
    return pool
}

func (p *DriverPool) Acquire(ctx context.Context) (*Driver, error) {
    select {
    case d := <-p.idleChan:
        return d, nil
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}

func (p *DriverPool) Release(d *Driver) {
    select {
    case p.idleChan <- d:
    default:
        // Пул полон, закрыть драйвер
        _ = d.Close()
    }
}
```

**8.7. Интеграция пула в ParseWorker**
```go
// internal/queue/worker_pools.go
type ParserPool struct {
    workers    int
    taskChan   chan *ParseTask
    handler    *ParseWorker
    ctx        context.Context
    wg         sync.WaitGroup
    mu         sync.RWMutex
    isRunning  bool
    driverPool *selenium.DriverPool  // ← пул драйверов
}

func NewParserPool(
    workers int,
    // ...
    cfg *config.ParserConfig,
) *ParserPool {
    // Пул Selenium-драйверов: 1 на каждого воркера
    driverPool := selenium.NewDriverPool(selenium.Config{
        Host:            cfg.SeleniumHost,
        Port:            cfg.SeleniumPort,
        Headless:        cfg.SeleniumHeadless,
        PageLoadTimeout: time.Duration(cfg.PageLoadTimeout) * time.Second,
    }, workers)
    
    return &ParserPool{
        workers:    workers,
        taskChan:   make(chan *ParseTask, workers*2),
        handler:    NewParseWorker(/* ... */, driverPool),
        driverPool: driverPool,
    }
}
```

---

## 9. Сводный план работ

| № | Задача | Файлы | Миграция |
|---|--------|-------|----------|
| 1 | Лимит/offset/priority из запроса | dto/parse.go, service/parse_service.go, handler/parse_handler.go | - |
| 2 | Несколько media_type в запросе | entity/request.go, repository/request_repo.go, service/request_service.go | 011_request_media_types.sql |
| 3 | Проверка статуса источника | service/source_service.go | 012_add_source_check_fields.sql |
| 4 | Удалить Download, создать endpoint | dto/*.go, service/*.go, handler/download_handler.go | - |
| 5 | TTL кэша из ENV | config/config.go, queue/parse_worker.go | - |
| 6 | Исправить статусы | entity/dictionary.go, queue/parse_worker.go, repository/*.go | - |
| 7 | Приоритет и retry | queue/producer.go, queue/parse_worker.go, repository/request_repo.go | 013_rabbitmq_priority_queue.sql |
| 8 | **Selenium: всегда использовать** | queue/parse_worker.go, parser/selenium/pool.go, queue/worker_pools.go | - |

---

## 10. Порядок выполнения

1. **Миграции БД** (011, 012, 013)
2. **Обновить entity и repository**
3. **Обновить service layer**
4. **Обновить handler layer**
5. **Обновить queue/worker**
6. **Тестирование**

---

## 11. Тестовые сценарии

### 11.1. Лимит и offset
```bash
# Запросить 10 медиа с offset 20
curl -X POST http://localhost:8080/api/v1/parse/batch \
  -H "X-Auth-Token: test" \
  -H "Content-Type: application/json" \
  -d '{"urls":["https://site.com"],"limit":10,"offset":20,"priority":5}'

# Проверить в БД:
SELECT limit_count, offset_count, priority FROM requests WHERE id = '...';
-- Ожидаем: 10, 20, 5
```

### 11.2. Несколько типов медиа
```bash
curl -X POST http://localhost:8080/api/v1/parse/batch \
  -H "X-Auth-Token: test" \
  -H "Content-Type: application/json" \
  -d '{"urls":["https://site.com"],"media_types":["image","video"],"limit":10}'

# Проверить:
SELECT * FROM request_media_types WHERE request_id = '...';
-- Ожидаем: 2 записи (image, video)
```

### 11.3. Проверка статуса источника
```bash
# Источник с 404
curl -X POST http://localhost:8080/api/v1/sources \
  -H "X-Auth-Token: test" \
  -d '{"name":"test","base_url":"https://nonexistent-domain-12345.com"}'

# Проверить:
SELECT status_id FROM sources WHERE base_url = '...';
-- Ожидаем: 3 (error)
```

### 11.4. TTL кэша
```bash
# Изменить в .env: CACHE_TTL_HOURS=1
# Перезапустить backend
# Парсить URL дважды с интервалом 2 часа
# Второй раз должно парситься заново
```

### 11.5. Приоритет
```bash
# Запрос с приоритетом 9
curl -X POST http://localhost:8080/api/v1/parse/url \
  -d '{"url":"https://site1.com","priority":9}'

# Запрос с приоритетом 1
curl -X POST http://localhost:8080/api/v1/parse/url \
  -d '{"url":"https://site2.com","priority":1}'

# Первый должен обработаться раньше
```

### 11.6. Selenium: парсинг JS-сайтов
```bash
# Pikabu (React) — должен парситься через Selenium
curl -X POST http://localhost:8080/api/v1/parse/url \
  -H "X-Auth-Token: test" \
  -d '{"url":"https://pikabu.ru"}'

# Проверить логи:
# ParseWorker: fetching URL https://pikabu.ru via Selenium
# ParseWorker: fetched 500000 bytes via Selenium

# Проверить найденные медиа:
SELECT COUNT(*) FROM media WHERE url LIKE '%pikabu.ru%';
-- Ожидаем: > 0 (изображения из постов)
```

### 11.7. Selenium: пул драйверов
```bash
# Запустить 10 запросов одновременно
for i in {1..10}; do
  curl -X POST http://localhost:8080/api/v1/parse/url \
    -H "X-Auth-Token: test" \
    -d "{\"url\":\"https://site$i.com\"}" &
done

# Проверить логи:
# ParserPool: started 5 workers
# DriverPool: created 5 drivers
# ParseWorker: using Selenium driver from pool
```
