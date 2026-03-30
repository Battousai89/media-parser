package dto

import (
	"time"

	"github.com/google/uuid"
)

type MediaItem struct {
	ID          uuid.UUID `json:"id"`
	URL         string    `json:"url"`
	MediaTypeID int       `json:"media_type_id"`
	MediaType   string    `json:"media_type"`
	Title       *string   `json:"title,omitempty"`
	FileSize    *int64    `json:"file_size,omitempty"`
	MimeType    *string   `json:"mime_type,omitempty"`
	Available   bool      `json:"available"`
	CreatedAt   time.Time `json:"created_at"`
}

type MediaDetail struct {
	ID          uuid.UUID     `json:"id"`
	URL         string        `json:"url"`
	MediaTypeID int           `json:"media_type_id"`
	MediaType   string        `json:"media_type"`
	Title       *string       `json:"title,omitempty"`
	Description *string       `json:"description,omitempty"`
	FileSize    *int64        `json:"file_size,omitempty"`
	MimeType    *string       `json:"mime_type,omitempty"`
	Hash        *string       `json:"hash,omitempty"`
	Meta        interface{}   `json:"meta,omitempty"`
	Available   bool          `json:"available"`
	CheckedAt   *time.Time    `json:"checked_at,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	Sources     []SourceBrief `json:"sources,omitempty"`
}

type SourceBrief struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	BaseURL string `json:"base_url"`
}

type MediaListRequest struct {
	MediaType *string `form:"media_type" binding:"omitempty,oneof=image video audio document archive other"`
	Available *bool   `form:"available" binding:"omitempty"`
	SourceID  *int    `form:"source_id" binding:"omitempty,min=1"`
	RequestID *string `form:"request_id" binding:"omitempty,uuid"`
	Search    *string `form:"search" binding:"omitempty,max=255"`
	Limit     int     `form:"limit" binding:"omitempty,min=1,max=100"`
	Offset    int     `form:"offset" binding:"omitempty,min=0"`
}

type MediaUploadRequest struct {
	MediaID  *string `json:"media_id" binding:"omitempty,uuid"`
	SourceID *int    `json:"source_id" binding:"omitempty,min=1"`
}

type MediaCheckRequest struct {
	URL string `json:"url" binding:"required,url"`
}

type MediaCheckResponse struct {
	URL       string  `json:"url"`
	Available bool    `json:"available"`
	FileSize  *int64  `json:"file_size,omitempty"`
	MimeType  *string `json:"mime_type,omitempty"`
}
