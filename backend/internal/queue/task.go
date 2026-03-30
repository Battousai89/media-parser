package queue

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type TaskType string

const (
	TaskTypeParse    TaskType = "parse"
	TaskTypeHTTP     TaskType = "http"
	TaskTypeDownload TaskType = "download"
)

type Task interface {
	GetType() TaskType
	GetRequestID() uuid.UUID
	GetSourceID() int
	GetURL() string
	GetCreatedAt() time.Time
}

type ParseTask struct {
	RequestID  uuid.UUID `json:"request_id"`
	SourceID   int       `json:"source_id"`
	URL        string    `json:"url"`
	MediaType  *string   `json:"media_type,omitempty"`
	Limit      int       `json:"limit"`
	Offset     int       `json:"offset"`
	Priority   int       `json:"priority"`
	RetryCount int       `json:"retry_count"`
	MaxRetries int       `json:"max_retries"`
	CreatedAt  time.Time `json:"created_at"`
}

func (t *ParseTask) GetType() TaskType       { return TaskTypeParse }
func (t *ParseTask) GetRequestID() uuid.UUID { return t.RequestID }
func (t *ParseTask) GetSourceID() int        { return t.SourceID }
func (t *ParseTask) GetURL() string          { return t.URL }
func (t *ParseTask) GetCreatedAt() time.Time { return t.CreatedAt }

type HTTPFetchTask struct {
	RequestID  uuid.UUID `json:"request_id"`
	SourceID   int       `json:"source_id"`
	URL        string    `json:"url"`
	MediaType  *string   `json:"media_type,omitempty"`
	Limit      int       `json:"limit"`
	ParentTask *ParseTask `json:"-"`
	CreatedAt  time.Time `json:"created_at"`
}

func (t *HTTPFetchTask) GetType() TaskType       { return TaskTypeHTTP }
func (t *HTTPFetchTask) GetRequestID() uuid.UUID { return t.RequestID }
func (t *HTTPFetchTask) GetSourceID() int        { return t.SourceID }
func (t *HTTPFetchTask) GetURL() string          { return t.URL }
func (t *HTTPFetchTask) GetCreatedAt() time.Time { return t.CreatedAt }

type DownloadTask struct {
	RequestID  uuid.UUID `json:"request_id"`
	SourceID   int       `json:"source_id"`
	MediaURL   string    `json:"media_url"`
	MediaType  string    `json:"media_type"`
	ParentTask *ParseTask `json:"-"`
	CreatedAt  time.Time `json:"created_at"`
}

func (t *DownloadTask) GetType() TaskType       { return TaskTypeDownload }
func (t *DownloadTask) GetRequestID() uuid.UUID { return t.RequestID }
func (t *DownloadTask) GetSourceID() int        { return t.SourceID }
func (t *DownloadTask) GetURL() string          { return t.MediaURL }
func (t *DownloadTask) GetCreatedAt() time.Time { return t.CreatedAt }

type ParsedMediaResult struct {
	RequestID uuid.UUID
	SourceID  int
	MediaURLs []MediaURL
	Error     error
	Duration  time.Duration
}

type MediaURL struct {
	URL       string
	MediaType string
	Title     *string
}

type HTTPFetchResult struct {
	RequestID  uuid.UUID
	SourceID   int
	URL        string
	HTML       string
	ContentType string
	StatusCode int
	Error      error
	Duration   time.Duration
	ParentTask *ParseTask
}

type DownloadResultInfo struct {
	RequestID uuid.UUID
	SourceID  int
	MediaURL  string
	FilePath  string
	FileSize  int64
	MimeType  string
	Error     error
	Duration  time.Duration
}

func MarshalTask(task Task) ([]byte, error) {
	return json.Marshal(task)
}

func UnmarshalParseTask(data []byte) (*ParseTask, error) {
	var task ParseTask
	err := json.Unmarshal(data, &task)
	return &task, err
}

func UnmarshalHTTPFetchTask(data []byte) (*HTTPFetchTask, error) {
	var task HTTPFetchTask
	err := json.Unmarshal(data, &task)
	return &task, err
}

func UnmarshalDownloadTask(data []byte) (*DownloadTask, error) {
	var task DownloadTask
	err := json.Unmarshal(data, &task)
	return &task, err
}

type TaskHandler func(ctx context.Context, task Task) error
type ParseTaskHandler func(ctx context.Context, task *ParseTask) error
type HTTPFetchHandler func(ctx context.Context, task *HTTPFetchTask) error
type DownloadHandler func(ctx context.Context, task *DownloadTask) error
