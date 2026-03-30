package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/media-parser/backend/internal/model/entity"
)

type RequestSourceRepositoryImpl struct {
	db *PostgresDB
}

func NewRequestSourceRepository(db *PostgresDB) *RequestSourceRepositoryImpl {
	return &RequestSourceRepositoryImpl{db: db}
}

func (r *RequestSourceRepositoryImpl) Create(ctx context.Context, rs *entity.RequestSource) error {
	query := `
		INSERT INTO request_sources (
			request_id, source_id, status_id, media_count, parsed_count,
			retry_count, max_retries, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id
	`
	return r.db.Pool.QueryRow(ctx, query,
		rs.RequestID, rs.SourceID, rs.StatusID, rs.MediaCount, rs.ParsedCount,
		rs.RetryCount, rs.MaxRetries,
	).Scan(&rs.ID)
}

func (r *RequestSourceRepositoryImpl) GetByRequestID(ctx context.Context, requestID uuid.UUID) ([]*entity.RequestSource, error) {
	query := `
		SELECT rs.id, rs.request_id, rs.source_id, rs.status_id, rs.media_count,
		       rs.parsed_count, rs.error_message, rs.retry_count, rs.max_retries,
		       rs.created_at, rs.updated_at,
		       s.name as source_name, s.base_url
		FROM request_sources rs
		LEFT JOIN sources s ON rs.source_id = s.id
		WHERE rs.request_id = $1
		ORDER BY rs.id
	`
	rows, err := r.db.Pool.Query(ctx, query, requestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []*entity.RequestSource
	for rows.Next() {
		rs := &entity.RequestSource{}
		var sourceName, baseURL string
		err := rows.Scan(
			&rs.ID, &rs.RequestID, &rs.SourceID, &rs.StatusID, &rs.MediaCount,
			&rs.ParsedCount, &rs.ErrorMessage, &rs.RetryCount, &rs.MaxRetries,
			&rs.CreatedAt, &rs.UpdatedAt, &sourceName, &baseURL,
		)
		if err != nil {
			return nil, err
		}
		// Создаём объект Source с именем и URL
		rs.Source = &entity.Source{
			ID:      rs.SourceID,
			Name:    sourceName,
			BaseURL: baseURL,
		}
		sources = append(sources, rs)
	}
	return sources, rows.Err()
}

func (r *RequestSourceRepositoryImpl) Update(ctx context.Context, rs *entity.RequestSource) error {
	query := `
		UPDATE request_sources
		SET status_id = $1, media_count = $2, parsed_count = $3,
		    error_message = $4, updated_at = NOW()
		WHERE request_id = $5 AND source_id = $6
	`
	_, err := r.db.Pool.Exec(ctx, query,
		rs.StatusID, rs.MediaCount, rs.ParsedCount, rs.ErrorMessage,
		rs.RequestID, rs.SourceID,
	)
	return err
}

func (r *RequestSourceRepositoryImpl) UpdateStatus(ctx context.Context, requestID uuid.UUID, sourceID int, statusID int, parsedCount int, errorMsg *string) error {
	query := `
		UPDATE request_sources
		SET status_id = $1, parsed_count = $2, error_message = $3, updated_at = NOW()
		WHERE request_id = $4 AND source_id = $5
	`
	_, err := r.db.Pool.Exec(ctx, query, statusID, parsedCount, errorMsg, requestID, sourceID)
	return err
}

func (r *RequestSourceRepositoryImpl) IncrementRetryCount(ctx context.Context, requestID uuid.UUID, sourceID int) error {
	query := `
		UPDATE request_sources
		SET retry_count = retry_count + 1, updated_at = NOW()
		WHERE request_id = $1 AND source_id = $2
	`
	_, err := r.db.Pool.Exec(ctx, query, requestID, sourceID)
	return err
}

func (r *RequestSourceRepositoryImpl) UpdateParsedCount(ctx context.Context, requestID uuid.UUID, sourceID int, parsedCount int) error {
	query := `
		UPDATE request_sources
		SET parsed_count = $1, updated_at = NOW()
		WHERE request_id = $2 AND source_id = $3
	`
	_, err := r.db.Pool.Exec(ctx, query, parsedCount, requestID, sourceID)
	return err
}

func (r *RequestSourceRepositoryImpl) GetByRequestIDAndSourceID(ctx context.Context, requestID uuid.UUID, sourceID int) (*entity.RequestSource, error) {
	query := `
		SELECT rs.id, rs.request_id, rs.source_id, rs.status_id, rs.media_count,
		       rs.parsed_count, rs.error_message, rs.retry_count, rs.max_retries,
		       rs.created_at, rs.updated_at,
		       s.name as source_name, s.base_url
		FROM request_sources rs
		LEFT JOIN sources s ON rs.source_id = s.id
		WHERE rs.request_id = $1 AND rs.source_id = $2
	`
	rs := &entity.RequestSource{}
	var sourceName, baseURL string
	err := r.db.Pool.QueryRow(ctx, query, requestID, sourceID).Scan(
		&rs.ID, &rs.RequestID, &rs.SourceID, &rs.StatusID, &rs.MediaCount,
		&rs.ParsedCount, &rs.ErrorMessage, &rs.RetryCount, &rs.MaxRetries,
		&rs.CreatedAt, &rs.UpdatedAt, &sourceName, &baseURL,
	)
	if err != nil {
		return nil, err
	}
	rs.Source = &entity.Source{
		ID:      rs.SourceID,
		Name:    sourceName,
		BaseURL: baseURL,
	}
	return rs, nil
}
