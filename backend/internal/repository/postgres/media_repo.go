package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/media-parser/backend/internal/model/entity"
)

type MediaRepositoryImpl struct {
	db *PostgresDB
}

func NewMediaRepository(db *PostgresDB) *MediaRepositoryImpl {
	return &MediaRepositoryImpl{db: db}
}

func (r *MediaRepositoryImpl) Create(ctx context.Context, media *entity.Media) error {
	query := `
		INSERT INTO media (
			id, url, media_type_id, title, description, file_size,
			mime_type, hash, storage_path, meta, available, checked_at, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW(), NOW())
		ON CONFLICT (url) DO UPDATE SET
			media_type_id = EXCLUDED.media_type_id,
			title = COALESCE(EXCLUDED.title, media.title),
			storage_path = EXCLUDED.storage_path,
			updated_at = NOW()
		RETURNING id
	`
	return r.db.Pool.QueryRow(ctx, query,
		media.ID, media.URL, media.MediaTypeID, media.Title, media.Description,
		media.FileSize, media.MimeType, media.Hash, media.StoragePath, media.Meta, media.Available, media.CheckedAt,
	).Scan(&media.ID)
}

func (r *MediaRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entity.Media, error) {
	query := `
		SELECT id, url, media_type_id, title, description, file_size,
		       mime_type, hash, storage_path, meta, available, checked_at, created_at, updated_at
		FROM media
		WHERE id = $1
	`
	media := &entity.Media{}
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&media.ID, &media.URL, &media.MediaTypeID, &media.Title, &media.Description,
		&media.FileSize, &media.MimeType, &media.Hash, &media.StoragePath, &media.Meta, &media.Available,
		&media.CheckedAt, &media.CreatedAt, &media.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return media, nil
}

func (r *MediaRepositoryImpl) GetByURL(ctx context.Context, url string) (*entity.Media, error) {
	query := `
		SELECT id, url, media_type_id, title, description, file_size,
		       mime_type, hash, storage_path, meta, available, checked_at, created_at, updated_at
		FROM media
		WHERE url = $1
	`
	media := &entity.Media{}
	err := r.db.Pool.QueryRow(ctx, query, url).Scan(
		&media.ID, &media.URL, &media.MediaTypeID, &media.Title, &media.Description,
		&media.FileSize, &media.MimeType, &media.Hash, &media.StoragePath, &media.Meta, &media.Available,
		&media.CheckedAt, &media.CreatedAt, &media.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return media, nil
}

func (r *MediaRepositoryImpl) GetAll(ctx context.Context, limit, offset int) ([]*entity.Media, error) {
	query := `
		SELECT id, url, media_type_id, title, description, file_size,
		       mime_type, hash, storage_path, meta, available, checked_at, created_at, updated_at
		FROM media
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var medias []*entity.Media
	for rows.Next() {
		media := &entity.Media{}
		err := rows.Scan(
			&media.ID, &media.URL, &media.MediaTypeID, &media.Title, &media.Description,
			&media.FileSize, &media.MimeType, &media.Hash, &media.StoragePath, &media.Meta, &media.Available,
			&media.CheckedAt, &media.CreatedAt, &media.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		medias = append(medias, media)
	}
	return medias, rows.Err()
}

func (r *MediaRepositoryImpl) GetByMediaType(ctx context.Context, mediaTypeID int, limit, offset int) ([]*entity.Media, error) {
	query := `
		SELECT id, url, media_type_id, title, description, file_size,
		       mime_type, hash, storage_path, meta, available, checked_at, created_at, updated_at
		FROM media
		WHERE media_type_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Pool.Query(ctx, query, mediaTypeID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var medias []*entity.Media
	for rows.Next() {
		media := &entity.Media{}
		err := rows.Scan(
			&media.ID, &media.URL, &media.MediaTypeID, &media.Title, &media.Description,
			&media.FileSize, &media.MimeType, &media.Hash, &media.StoragePath, &media.Meta, &media.Available,
			&media.CheckedAt, &media.CreatedAt, &media.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		medias = append(medias, media)
	}
	return medias, rows.Err()
}

func (r *MediaRepositoryImpl) GetByHash(ctx context.Context, hash string) (*entity.Media, error) {
	query := `
		SELECT id, url, media_type_id, title, description, file_size,
		       mime_type, hash, storage_path, meta, available, checked_at, created_at, updated_at
		FROM media
		WHERE hash = $1
	`
	media := &entity.Media{}
	err := r.db.Pool.QueryRow(ctx, query, hash).Scan(
		&media.ID, &media.URL, &media.MediaTypeID, &media.Title, &media.Description,
		&media.FileSize, &media.MimeType, &media.Hash, &media.StoragePath, &media.Meta, &media.Available,
		&media.CheckedAt, &media.CreatedAt, &media.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return media, nil
}

func (r *MediaRepositoryImpl) Update(ctx context.Context, media *entity.Media) error {
	query := `
		UPDATE media
		SET title = $1, description = $2, file_size = $3, mime_type = $4,
		    hash = $5, storage_path = $6, meta = $7, available = $8, checked_at = $9, updated_at = NOW()
		WHERE id = $10
	`
	_, err := r.db.Pool.Exec(ctx, query,
		media.Title, media.Description, media.FileSize, media.MimeType,
		media.Hash, media.StoragePath, media.Meta, media.Available, media.CheckedAt, media.ID,
	)
	return err
}

func (r *MediaRepositoryImpl) UpdateStoragePath(ctx context.Context, id uuid.UUID, storagePath string) error {
	query := `
		UPDATE media
		SET storage_path = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.Pool.Exec(ctx, query, storagePath, id)
	return err
}

func (r *MediaRepositoryImpl) UpdateAvailability(ctx context.Context, id uuid.UUID, available bool) error {
	query := `
		UPDATE media
		SET available = $1, checked_at = NOW(), updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.Pool.Exec(ctx, query, available, id)
	return err
}

func (r *MediaRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM media WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id)
	return err
}

func (r *MediaRepositoryImpl) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM media`).Scan(&count)
	return count, err
}

func (r *MediaRepositoryImpl) CountByMediaType(ctx context.Context, mediaTypeID int) (int, error) {
	var count int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM media WHERE media_type_id = $1`, mediaTypeID).Scan(&count)
	return count, err
}
