package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Media struct {
	ID          uuid.UUID       `db:"id" json:"id"`
	URL         string          `db:"url" json:"url"`
	MediaTypeID int             `db:"media_type_id" json:"media_type_id"`
	MediaType   *MediaType      `db:"-" json:"media_type,omitempty"`
	Title       *string         `db:"title" json:"title,omitempty"`
	Description *string         `db:"description" json:"description,omitempty"`
	FileSize    *int64          `db:"file_size" json:"file_size,omitempty"`
	MimeType    *string         `db:"mime_type" json:"mime_type,omitempty"`
	Hash        *string         `db:"hash" json:"hash,omitempty"`
	StoragePath *string         `db:"storage_path" json:"storage_path,omitempty"`
	Meta        json.RawMessage `db:"meta" json:"meta,omitempty"`
	Available   bool            `db:"available" json:"available"`
	CheckedAt   *time.Time      `db:"checked_at" json:"checked_at,omitempty"`
	CreatedAt   time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time       `db:"updated_at" json:"updated_at"`
}

type SourceMedia struct {
	ID        int        `db:"id" json:"id"`
	SourceID  int        `db:"source_id" json:"source_id"`
	Source    *Source    `db:"-" json:"source,omitempty"`
	MediaID   uuid.UUID  `db:"media_id" json:"media_id"`
	Media     *Media     `db:"-" json:"media,omitempty"`
	RequestID *uuid.UUID `db:"request_id" json:"request_id,omitempty"`
	Request   *Request   `db:"-" json:"request,omitempty"`
	FoundAt   time.Time  `db:"found_at" json:"found_at"`
}
