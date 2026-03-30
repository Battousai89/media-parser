package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/media-parser/backend/internal/model/entity"
)

type SourceMediaRepositoryImpl struct {
	db *PostgresDB
}

func NewSourceMediaRepository(db *PostgresDB) *SourceMediaRepositoryImpl {
	return &SourceMediaRepositoryImpl{db: db}
}

func (r *SourceMediaRepositoryImpl) Create(ctx context.Context, sm *entity.SourceMedia) error {
	query := `
		INSERT INTO source_media (source_id, media_id, request_id, found_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (source_id, media_id) DO UPDATE
		SET request_id = $3, found_at = NOW()
		RETURNING id
	`
	return r.db.Pool.QueryRow(ctx, query, sm.SourceID, sm.MediaID, sm.RequestID).Scan(&sm.ID)
}

func (r *SourceMediaRepositoryImpl) GetBySourceID(ctx context.Context, sourceID int, limit, offset int) ([]*entity.SourceMedia, error) {
	query := `
		SELECT id, source_id, media_id, request_id, found_at
		FROM source_media
		WHERE source_id = $1
		ORDER BY found_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Pool.Query(ctx, query, sourceID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sourceMedias []*entity.SourceMedia
	for rows.Next() {
		sm := &entity.SourceMedia{}
		err := rows.Scan(&sm.ID, &sm.SourceID, &sm.MediaID, &sm.RequestID, &sm.FoundAt)
		if err != nil {
			return nil, err
		}
		sourceMedias = append(sourceMedias, sm)
	}
	return sourceMedias, rows.Err()
}

func (r *SourceMediaRepositoryImpl) GetByRequestID(ctx context.Context, requestID uuid.UUID, limit, offset int) ([]*entity.SourceMedia, error) {
	query := `
		SELECT id, source_id, media_id, request_id, found_at
		FROM source_media
		WHERE request_id = $1
		ORDER BY found_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Pool.Query(ctx, query, requestID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sourceMedias []*entity.SourceMedia
	for rows.Next() {
		sm := &entity.SourceMedia{}
		err := rows.Scan(&sm.ID, &sm.SourceID, &sm.MediaID, &sm.RequestID, &sm.FoundAt)
		if err != nil {
			return nil, err
		}
		sourceMedias = append(sourceMedias, sm)
	}
	return sourceMedias, rows.Err()
}

func (r *SourceMediaRepositoryImpl) GetByMediaID(ctx context.Context, mediaID uuid.UUID) ([]*entity.SourceMedia, error) {
	query := `
		SELECT id, source_id, media_id, request_id, found_at
		FROM source_media
		WHERE media_id = $1
	`
	rows, err := r.db.Pool.Query(ctx, query, mediaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sourceMedias []*entity.SourceMedia
	for rows.Next() {
		sm := &entity.SourceMedia{}
		err := rows.Scan(&sm.ID, &sm.SourceID, &sm.MediaID, &sm.RequestID, &sm.FoundAt)
		if err != nil {
			return nil, err
		}
		sourceMedias = append(sourceMedias, sm)
	}
	return sourceMedias, rows.Err()
}
