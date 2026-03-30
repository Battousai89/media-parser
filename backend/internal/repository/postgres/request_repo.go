package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/media-parser/backend/internal/model/dto"
	"github.com/media-parser/backend/internal/model/entity"
)

type RequestRepositoryImpl struct {
	db *PostgresDB
}

func NewRequestRepository(db *PostgresDB) *RequestRepositoryImpl {
	return &RequestRepositoryImpl{db: db}
}

func (r *RequestRepositoryImpl) Create(ctx context.Context, req *entity.Request) error {
	query := `
		INSERT INTO requests (
			id, status_id, limit_count, offset_count,
			priority, retry_count, max_retries, token_id, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
	`
	_, err := r.db.Pool.Exec(ctx, query,
		req.ID, req.StatusID, req.LimitCount, req.OffsetCount,
		req.Priority, req.RetryCount, req.MaxRetries, req.TokenID,
	)
	return err
}

func (r *RequestRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entity.Request, error) {
	query := `
		SELECT r.id, r.status_id, rs.code, r.limit_count, r.offset_count,
		       r.priority, r.retry_count, r.max_retries, r.error_message,
		       r.started_at, r.completed_at, r.created_at, r.updated_at
		FROM requests r
		LEFT JOIN request_statuses rs ON r.status_id = rs.id
		WHERE r.id = $1
	`
	req := &entity.Request{}
	var statusCode string
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&req.ID, &req.StatusID, &statusCode, &req.LimitCount, &req.OffsetCount,
		&req.Priority, &req.RetryCount, &req.MaxRetries, &req.ErrorMessage,
		&req.StartedAt, &req.CompletedAt, &req.CreatedAt, &req.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	req.Status = &entity.RequestStatus{ID: req.StatusID, Code: statusCode}

	mediaTypeIDs, err := r.GetMediaTypeIDs(ctx, id)
	if err == nil && len(mediaTypeIDs) > 0 {
		req.MediaTypeIDs = mediaTypeIDs
	}

	return req, nil
}

func (r *RequestRepositoryImpl) GetAll(ctx context.Context, limit, offset int) ([]*entity.Request, error) {
	query := `
		SELECT r.id, r.status_id, rs.code, r.limit_count, r.offset_count,
		       r.priority, r.retry_count, r.max_retries, r.error_message,
		       r.started_at, r.completed_at, r.created_at, r.updated_at
		FROM requests r
		LEFT JOIN request_statuses rs ON r.status_id = rs.id
		ORDER BY r.created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*entity.Request
	for rows.Next() {
		req := &entity.Request{}
		var statusCode string
		err := rows.Scan(
			&req.ID, &req.StatusID, &statusCode, &req.LimitCount, &req.OffsetCount,
			&req.Priority, &req.RetryCount, &req.MaxRetries, &req.ErrorMessage,
			&req.StartedAt, &req.CompletedAt, &req.CreatedAt, &req.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		req.Status = &entity.RequestStatus{ID: req.StatusID, Code: statusCode}
		requests = append(requests, req)
	}
	return requests, rows.Err()
}

func (r *RequestRepositoryImpl) GetByStatus(ctx context.Context, statusID int, limit, offset int) ([]*entity.Request, error) {
	query := `
		SELECT r.id, r.status_id, rs.code, r.limit_count, r.offset_count,
		       r.priority, r.retry_count, r.max_retries, r.error_message,
		       r.started_at, r.completed_at, r.created_at, r.updated_at
		FROM requests r
		LEFT JOIN request_statuses rs ON r.status_id = rs.id
		WHERE r.status_id = $1
		ORDER BY r.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Pool.Query(ctx, query, statusID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*entity.Request
	for rows.Next() {
		req := &entity.Request{}
		var statusCode string
		err := rows.Scan(
			&req.ID, &req.StatusID, &statusCode, &req.LimitCount, &req.OffsetCount,
			&req.Priority, &req.RetryCount, &req.MaxRetries, &req.ErrorMessage,
			&req.StartedAt, &req.CompletedAt, &req.CreatedAt, &req.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		req.Status = &entity.RequestStatus{ID: req.StatusID, Code: statusCode}
		requests = append(requests, req)
	}
	return requests, rows.Err()
}

func (r *RequestRepositoryImpl) Update(ctx context.Context, req *entity.Request) error {
	query := `
		UPDATE requests
		SET status_id = $1, limit_count = $2, offset_count = $3,
		    priority = $4, retry_count = $5, max_retries = $6, error_message = $7,
		    started_at = $8, completed_at = $9, updated_at = NOW()
		WHERE id = $10
	`
	_, err := r.db.Pool.Exec(ctx, query,
		req.StatusID, req.LimitCount, req.OffsetCount,
		req.Priority, req.RetryCount, req.MaxRetries, req.ErrorMessage,
		req.StartedAt, req.CompletedAt, req.ID,
	)
	return err
}

func (r *RequestRepositoryImpl) UpdateStatus(ctx context.Context, id uuid.UUID, statusID int, errorMsg *string) error {
	query := `
		UPDATE requests
		SET status_id = $1, error_message = $2, updated_at = NOW(),
		    started_at = CASE WHEN status_id = 1 THEN NOW() ELSE started_at END,
		    completed_at = CASE WHEN $1 IN (3, 4, 5) THEN NOW() ELSE completed_at END
		WHERE id = $3
	`
	_, err := r.db.Pool.Exec(ctx, query, statusID, errorMsg, id)
	return err
}

func (r *RequestRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM requests WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id)
	return err
}

func (r *RequestRepositoryImpl) GetRequestMedia(ctx context.Context, requestID uuid.UUID, limit, offset int) ([]*entity.Media, error) {
	query := `
		SELECT DISTINCT ON (m.id) m.id, m.url, m.media_type_id, mt.code, m.title, m.description,
		       m.file_size, m.mime_type, m.hash, m.meta, m.available, m.checked_at, m.created_at, m.updated_at
		FROM media m
		INNER JOIN source_media sm ON m.id = sm.media_id
		LEFT JOIN media_types mt ON m.media_type_id = mt.id
		WHERE sm.request_id = $1
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Pool.Query(ctx, query, requestID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var medias []*entity.Media
	for rows.Next() {
		m := &entity.Media{}
		var mediaTypeCode string
		var mediaTypeID int
		err := rows.Scan(
			&m.ID, &m.URL, &mediaTypeID, &mediaTypeCode, &m.Title, &m.Description,
			&m.FileSize, &m.MimeType, &m.Hash, &m.Meta, &m.Available, &m.CheckedAt, &m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		m.MediaTypeID = mediaTypeID
		m.MediaType = &entity.MediaType{ID: mediaTypeID, Code: mediaTypeCode}
		medias = append(medias, m)
	}
	return medias, rows.Err()
}

func (r *RequestRepositoryImpl) GetRequestMediaWithSources(ctx context.Context, requestID uuid.UUID, limit, offset int) ([]*dto.RequestMediaItem, error) {
	query := `
		SELECT m.id, m.url, m.media_type_id, COALESCE(m.title, ''), COALESCE(m.file_size, 0), COALESCE(m.mime_type, ''),
		       sm.source_id, m.available, m.created_at
		FROM media m
		INNER JOIN source_media sm ON m.id = sm.media_id
		WHERE sm.request_id = $1
		ORDER BY sm.source_id, m.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Pool.Query(ctx, query, requestID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var medias []*dto.RequestMediaItem
	for rows.Next() {
		var id uuid.UUID
		var url string
		var mediaTypeID int
		var title string
		var fileSize int64
		var mimeType string
		var sourceID int
		var available bool
		var createdAt time.Time
		err := rows.Scan(&id, &url, &mediaTypeID, &title, &fileSize, &mimeType, &sourceID, &available, &createdAt)
		if err != nil {
			return nil, err
		}
		m := &dto.RequestMediaItem{
			ID:          id,
			URL:         url,
			MediaTypeID: mediaTypeID,
			SourceID:    sourceID,
			Available:   available,
			CreatedAt:   createdAt,
		}
		if title != "" {
			m.Title = &title
		}
		if fileSize > 0 {
			m.FileSize = &fileSize
		}
		if mimeType != "" {
			m.MimeType = &mimeType
		}
		medias = append(medias, m)
	}
	return medias, rows.Err()
}

func (r *RequestRepositoryImpl) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM requests`).Scan(&count)
	return count, err
}

func (r *RequestRepositoryImpl) IncrementRetryCount(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE requests
		SET retry_count = retry_count + 1, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.Pool.Exec(ctx, query, id)
	return err
}

func (r *RequestRepositoryImpl) GetMediaTypeIDs(ctx context.Context, requestID uuid.UUID) ([]int, error) {
	query := `
		SELECT media_type_id FROM request_media_types
		WHERE request_id = $1
		ORDER BY media_type_id
	`
	rows, err := r.db.Pool.Query(ctx, query, requestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mediaTypeIDs []int
	for rows.Next() {
		var mediaTypeID int
		err := rows.Scan(&mediaTypeID)
		if err != nil {
			return nil, err
		}
		mediaTypeIDs = append(mediaTypeIDs, mediaTypeID)
	}
	return mediaTypeIDs, rows.Err()
}

func (r *RequestRepositoryImpl) SetMediaTypeIDs(ctx context.Context, requestID uuid.UUID, mediaTypeIDs []int) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	deleteQuery := `DELETE FROM request_media_types WHERE request_id = $1`
	_, err = tx.Exec(ctx, deleteQuery, requestID)
	if err != nil {
		return err
	}

	insertQuery := `
		INSERT INTO request_media_types (request_id, media_type_id, created_at)
		VALUES ($1, $2, NOW())
	`
	for _, mediaTypeID := range mediaTypeIDs {
		_, err = tx.Exec(ctx, insertQuery, requestID, mediaTypeID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
