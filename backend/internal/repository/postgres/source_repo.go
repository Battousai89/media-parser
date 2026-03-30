package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/media-parser/backend/internal/model/entity"
)

type SourceRepositoryImpl struct {
	db *PostgresDB
}

func NewSourceRepository(db *PostgresDB) *SourceRepositoryImpl {
	return &SourceRepositoryImpl{db: db}
}

func (r *SourceRepositoryImpl) Create(ctx context.Context, source *entity.Source) error {
	query := `
		INSERT INTO sources (name, base_url, status_id, updated_at, created_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id
	`
	return r.db.Pool.QueryRow(ctx, query, source.Name, source.BaseURL, source.StatusID).Scan(&source.ID)
}

func (r *SourceRepositoryImpl) GetByID(ctx context.Context, id int) (*entity.Source, error) {
	query := `
		SELECT id, name, base_url, status_id, last_checked_at, check_error_message, updated_at, created_at
		FROM sources
		WHERE id = $1
	`
	source := &entity.Source{}
	var lastCheckedAt interface{}
	var checkErrorMessage interface{}
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&source.ID, &source.Name, &source.BaseURL, &source.StatusID,
		&lastCheckedAt, &checkErrorMessage,
		&source.UpdatedAt, &source.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if t, ok := lastCheckedAt.(time.Time); ok {
		source.LastCheckedAt = &t
	}
	if s, ok := checkErrorMessage.(string); ok {
		source.CheckErrorMessage = &s
	}
	return source, nil
}

func (r *SourceRepositoryImpl) GetAll(ctx context.Context) ([]*entity.Source, error) {
	query := `
		SELECT id, name, base_url, status_id, last_checked_at, check_error_message, updated_at, created_at
		FROM sources
		ORDER BY created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []*entity.Source
	for rows.Next() {
		source := &entity.Source{}
		var lastCheckedAt, checkErrorMessage interface{}
		err := rows.Scan(
			&source.ID, &source.Name, &source.BaseURL, &source.StatusID,
			&lastCheckedAt, &checkErrorMessage,
			&source.UpdatedAt, &source.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		if t, ok := lastCheckedAt.(time.Time); ok {
			source.LastCheckedAt = &t
		}
		if s, ok := checkErrorMessage.(string); ok {
			source.CheckErrorMessage = &s
		}
		sources = append(sources, source)
	}
	return sources, rows.Err()
}

func (r *SourceRepositoryImpl) GetActive(ctx context.Context) ([]*entity.Source, error) {
	query := `
		SELECT id, name, base_url, status_id, last_checked_at, check_error_message, updated_at, created_at
		FROM sources
		WHERE status_id = 1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []*entity.Source
	for rows.Next() {
		source := &entity.Source{}
		var lastCheckedAt, checkErrorMessage interface{}
		err := rows.Scan(
			&source.ID, &source.Name, &source.BaseURL, &source.StatusID,
			&lastCheckedAt, &checkErrorMessage,
			&source.UpdatedAt, &source.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		if t, ok := lastCheckedAt.(time.Time); ok {
			source.LastCheckedAt = &t
		}
		if s, ok := checkErrorMessage.(string); ok {
			source.CheckErrorMessage = &s
		}
		sources = append(sources, source)
	}
	return sources, rows.Err()
}

func (r *SourceRepositoryImpl) Update(ctx context.Context, source *entity.Source) error {
	query := `
		UPDATE sources
		SET name = $1, base_url = $2, status_id = $3, last_checked_at = $4, check_error_message = $5, updated_at = NOW()
		WHERE id = $6
	`
	_, err := r.db.Pool.Exec(ctx, query, source.Name, source.BaseURL, source.StatusID, source.LastCheckedAt, source.CheckErrorMessage, source.ID)
	return err
}

func (r *SourceRepositoryImpl) UpdateStatus(ctx context.Context, id int, statusID int) error {
	query := `
		UPDATE sources
		SET status_id = $1, last_checked_at = NOW(), updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.Pool.Exec(ctx, query, statusID, id)
	return err
}

func (r *SourceRepositoryImpl) UpdateStatusWithError(ctx context.Context, id int, statusID int, errorMessage *string) error {
	query := `
		UPDATE sources
		SET status_id = $1, check_error_message = $2, last_checked_at = NOW(), updated_at = NOW()
		WHERE id = $3
	`
	_, err := r.db.Pool.Exec(ctx, query, statusID, errorMessage, id)
	return err
}

func (r *SourceRepositoryImpl) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM sources WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id)
	return err
}

func (r *SourceRepositoryImpl) GetByURL(ctx context.Context, url string) (*entity.Source, error) {
	query := `
		SELECT id, name, base_url, status_id, last_checked_at, check_error_message, updated_at, created_at
		FROM sources
		WHERE base_url = $1
	`
	source := &entity.Source{}
	var lastCheckedAt, checkErrorMessage interface{}
	err := r.db.Pool.QueryRow(ctx, query, url).Scan(
		&source.ID, &source.Name, &source.BaseURL, &source.StatusID,
		&lastCheckedAt, &checkErrorMessage,
		&source.UpdatedAt, &source.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if t, ok := lastCheckedAt.(time.Time); ok {
		source.LastCheckedAt = &t
	}
	if s, ok := checkErrorMessage.(string); ok {
		source.CheckErrorMessage = &s
	}
	return source, nil
}
