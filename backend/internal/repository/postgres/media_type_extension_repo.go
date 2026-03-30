package postgres

import (
	"context"

	"github.com/media-parser/backend/internal/model/entity"
)

type MediaTypeExtensionRepository struct {
	db *PostgresDB
}

func NewMediaTypeExtensionRepository(db *PostgresDB) *MediaTypeExtensionRepository {
	return &MediaTypeExtensionRepository{db: db}
}

func (r *MediaTypeExtensionRepository) GetByMediaTypeID(ctx context.Context, mediaTypeID int) ([]*entity.MediaTypeExtension, error) {
	query := `SELECT id, media_type_id, extension FROM media_type_extensions WHERE media_type_id = $1 ORDER BY extension`
	rows, err := r.db.Pool.Query(ctx, query, mediaTypeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var extensions []*entity.MediaTypeExtension
	for rows.Next() {
		ext := &entity.MediaTypeExtension{}
		err := rows.Scan(&ext.ID, &ext.MediaTypeID, &ext.Extension)
		if err != nil {
			return nil, err
		}
		extensions = append(extensions, ext)
	}
	return extensions, rows.Err()
}

func (r *MediaTypeExtensionRepository) GetMediaTypeByExtension(ctx context.Context, extension string) (*int, error) {
	query := `SELECT media_type_id FROM media_type_extensions WHERE extension = $1 LIMIT 1`
	var mediaTypeID int
	err := r.db.Pool.QueryRow(ctx, query, extension).Scan(&mediaTypeID)
	if err != nil {
		return nil, err
	}
	return &mediaTypeID, nil
}
