package entity

import "time"

type Pattern struct {
	ID          int        `db:"id" json:"id"`
	Name        string     `db:"name" json:"name"`
	Regex       string     `db:"regex" json:"regex"`
	MediaTypeID int        `db:"media_type_id" json:"media_type_id"`
	MediaType   *MediaType `db:"-" json:"media_type,omitempty"`
	Priority    int        `db:"priority" json:"priority"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
}
