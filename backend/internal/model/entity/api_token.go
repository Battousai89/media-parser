package entity

import (
	"encoding/json"
	"time"
)

type APIToken struct {
	ID          int             `db:"id" json:"id"`
	Token       string          `db:"token" json:"token"`
	Name        *string         `db:"name" json:"name,omitempty"`
	Active      bool            `db:"active" json:"active"`
	ExpiresAt   *time.Time      `db:"expires_at" json:"expires_at,omitempty"`
	Permissions json.RawMessage `db:"permissions" json:"permissions,omitempty"`
	CreatedAt   time.Time       `db:"created_at" json:"created_at"`
	LastUsedAt  *time.Time      `db:"last_used_at" json:"last_used_at,omitempty"`
}

type TokenPermissions struct {
	Parse        bool `json:"parse"`
	MediaRead    bool `json:"media_read"`
	RequestsView bool `json:"requests_view"`
}

type URLCache struct {
	ID        int       `db:"id" json:"id"`
	URL       string    `db:"url" json:"url"`
	Hash      string    `db:"hash" json:"hash"`
	ParsedAt  time.Time `db:"parsed_at" json:"parsed_at"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
}
