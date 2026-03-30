package entity

import (
	"time"

	"github.com/google/uuid"
)

type Request struct {
	ID           uuid.UUID      `db:"id" json:"id"`
	StatusID     int            `db:"status_id" json:"status_id"`
	Status       *RequestStatus `db:"-" json:"status,omitempty"`
	MediaTypeIDs []int          `db:"-" json:"media_type_ids,omitempty"`
	LimitCount   int            `db:"limit_count" json:"limit_count"`
	OffsetCount  int            `db:"offset_count" json:"offset_count"`
	Priority     int            `db:"priority" json:"priority"`
	RetryCount   int            `db:"retry_count" json:"retry_count"`
	MaxRetries   int            `db:"max_retries" json:"max_retries"`
	ErrorMessage *string        `db:"error_message" json:"error_message,omitempty"`
	StartedAt    *time.Time     `db:"started_at" json:"started_at,omitempty"`
	CompletedAt  *time.Time     `db:"completed_at" json:"completed_at,omitempty"`
	CreatedAt    time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at" json:"updated_at"`
	TokenID      *int           `db:"token_id" json:"token_id,omitempty"`
}

type RequestSource struct {
	ID           int            `db:"id" json:"id"`
	RequestID    uuid.UUID      `db:"request_id" json:"request_id"`
	SourceID     int            `db:"source_id" json:"source_id"`
	Source       *Source        `db:"-" json:"source,omitempty"`
	StatusID     int            `db:"status_id" json:"status_id"`
	Status       *RequestStatus `db:"-" json:"status,omitempty"`
	MediaCount   int            `db:"media_count" json:"media_count"`
	ParsedCount  int            `db:"parsed_count" json:"parsed_count"`
	RetryCount   int            `db:"retry_count" json:"retry_count"`
	MaxRetries   int            `db:"max_retries" json:"max_retries"`
	ErrorMessage *string        `db:"error_message" json:"error_message,omitempty"`
	CreatedAt    time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at" json:"updated_at"`
}
