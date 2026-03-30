package dto

import "time"

type SourceItem struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	BaseURL   string    `json:"base_url"`
	StatusID  int       `json:"status_id"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SourceDetail struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	BaseURL   string    `json:"base_url"`
	StatusID  int       `json:"status_id"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

type SourceCreateRequest struct {
	Name    string `json:"name" binding:"required,max=255"`
	BaseURL string `json:"base_url" binding:"required,url,max=2048"`
	StatusID *int  `json:"status_id,omitempty" binding:"omitempty,min=1"`
}

type SourceUpdateRequest struct {
	Name    *string `json:"name,omitempty" binding:"omitempty,max=255"`
	BaseURL *string `json:"base_url,omitempty" binding:"omitempty,url,max=2048"`
	StatusID *int   `json:"status_id,omitempty" binding:"omitempty,min=1"`
}

type PatternItem struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Regex       string    `json:"regex"`
	MediaTypeID int       `json:"media_type_id"`
	MediaType   string    `json:"media_type"`
	Priority    int       `json:"priority"`
	CreatedAt   time.Time `json:"created_at"`
}

type PatternCreateRequest struct {
	Name        string `json:"name" binding:"required,max=255"`
	Regex       string `json:"regex" binding:"required"`
	MediaTypeID int    `json:"media_type_id" binding:"required,min=1"`
	Priority    *int   `json:"priority,omitempty" binding:"omitempty,min=0"`
}

type PatternUpdateRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,max=255"`
	Regex       *string `json:"regex,omitempty" binding:"omitempty"`
	MediaTypeID *int    `json:"media_type_id,omitempty" binding:"omitempty,min=1"`
	Priority    *int    `json:"priority,omitempty" binding:"omitempty,min=0"`
}
