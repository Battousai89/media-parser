package dto

import (
	"time"

	"github.com/google/uuid"
)

type RequestItem struct {
	ID           uuid.UUID `json:"id"`
	StatusID     int       `json:"status_id"`
	Status       string    `json:"status"`
	MediaTypeIDs []int     `json:"media_type_ids,omitempty"`
	LimitCount   int       `json:"limit_count"`
	OffsetCount  int       `json:"offset_count"`
	Priority     int       `json:"priority"`
	ParsedCount  int       `json:"parsed_count"`
	SourcesCount int       `json:"sources_count"`
	CreatedAt    time.Time `json:"created_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
}

type RequestDetail struct {
	ID           uuid.UUID           `json:"id"`
	StatusID     int                 `json:"status_id"`
	Status       string              `json:"status"`
	MediaTypeIDs []int               `json:"media_type_ids,omitempty"`
	LimitCount   int                 `json:"limit_count"`
	OffsetCount  int                 `json:"offset_count"`
	Priority     int                 `json:"priority"`
	RetryCount   int                 `json:"retry_count"`
	MaxRetries   int                 `json:"max_retries"`
	ErrorMessage *string             `json:"error_message,omitempty"`
	StartedAt    *time.Time          `json:"started_at,omitempty"`
	CompletedAt  *time.Time          `json:"completed_at,omitempty"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
	Sources      []RequestSourceItem `json:"sources,omitempty"`
}

type RequestSourceItem struct {
	SourceID     int     `json:"source_id"`
	SourceName   string  `json:"source_name"`
	BaseURL      string  `json:"base_url,omitempty"`
	StatusID     int     `json:"status_id"`
	Status       string  `json:"status"`
	MediaCount   int     `json:"media_count"`
	ParsedCount  int     `json:"parsed_count"`
	RetryCount   int     `json:"retry_count"`
	MaxRetries   int     `json:"max_retries"`
	ErrorMessage *string `json:"error_message,omitempty"`
}

type RequestMediaItem struct {
	ID          uuid.UUID `json:"id"`
	URL         string    `json:"url"`
	MediaTypeID int       `json:"media_type_id"`
	Title       *string   `json:"title,omitempty"`
	FileSize    *int64    `json:"file_size,omitempty"`
	MimeType    *string   `json:"mime_type,omitempty"`
	SourceID    int       `json:"source_id"`
	Available   bool      `json:"available"`
	CreatedAt   time.Time `json:"created_at"`
}

type RequestListRequest struct {
	StatusID  *int    `form:"status_id" binding:"omitempty,min=1"`
	MediaType *string `form:"media_type" binding:"omitempty,oneof=image video audio document archive other"`
	Limit     int     `form:"limit" binding:"omitempty,min=1,max=100"`
	Offset    int     `form:"offset" binding:"omitempty,min=0"`
}
