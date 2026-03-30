package entity

import "time"

type Source struct {
	ID             int           `db:"id" json:"id"`
	Name           string        `db:"name" json:"name"`
	BaseURL        string        `db:"base_url" json:"base_url"`
	StatusID       int           `db:"status_id" json:"status_id"`
	Status         *SourceStatus `db:"-" json:"status,omitempty"`
	LastCheckedAt  *time.Time    `db:"last_checked_at" json:"last_checked_at,omitempty"`
	CheckErrorMessage *string    `db:"check_error_message" json:"check_error_message,omitempty"`
	UpdatedAt      time.Time     `db:"updated_at" json:"updated_at"`
	CreatedAt      time.Time     `db:"created_at" json:"created_at"`
}
