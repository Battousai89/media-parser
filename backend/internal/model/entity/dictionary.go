package entity

import "time"

type RequestStatus struct {
	ID   int    `db:"id" json:"id"`
	Code string `db:"code" json:"code"`
	Name string `db:"name" json:"name"`
}

type MediaType struct {
	ID   int    `db:"id" json:"id"`
	Code string `db:"code" json:"code"`
	Name string `db:"name" json:"name"`
}

type MediaTypeExtension struct {
	ID          int    `db:"id" json:"id"`
	MediaTypeID int    `db:"media_type_id" json:"media_type_id"`
	Extension   string `db:"extension" json:"extension"`
}

type SourceStatus struct {
	ID   int    `db:"id" json:"id"`
	Code string `db:"code" json:"code"`
	Name string `db:"name" json:"name"`
}

// Status codes used in services
const (
	StatusPending = "pending"
)

// Source status codes used in services
const (
	SourceStatusActive   = "active"
	SourceStatusInactive = "inactive"
	SourceStatusError    = "error"
	SourceStatusBlocked  = "blocked"
)

type Timestamps struct {
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type NullableTimestamps struct {
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}
