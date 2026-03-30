package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/media-parser/backend/internal/model/entity"
)

type PatternRepositoryImpl struct {
	db *PostgresDB
}

func NewPatternRepository(db *PostgresDB) *PatternRepositoryImpl {
	return &PatternRepositoryImpl{db: db}
}

func (r *PatternRepositoryImpl) Create(ctx context.Context, pattern *entity.Pattern) error {
	query := `
		INSERT INTO patterns (name, regex, media_type_id, priority, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id
	`
	return r.db.Pool.QueryRow(ctx, query, pattern.Name, pattern.Regex, pattern.MediaTypeID, pattern.Priority).Scan(&pattern.ID)
}

func (r *PatternRepositoryImpl) GetByID(ctx context.Context, id int) (*entity.Pattern, error) {
	query := `
		SELECT id, name, regex, media_type_id, priority, created_at
		FROM patterns
		WHERE id = $1
	`
	pattern := &entity.Pattern{}
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&pattern.ID, &pattern.Name, &pattern.Regex, &pattern.MediaTypeID,
		&pattern.Priority, &pattern.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return pattern, nil
}

func (r *PatternRepositoryImpl) GetAll(ctx context.Context) ([]*entity.Pattern, error) {
	query := `
		SELECT id, name, regex, media_type_id, priority, created_at
		FROM patterns
		ORDER BY priority DESC, id
	`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var patterns []*entity.Pattern
	for rows.Next() {
		pattern := &entity.Pattern{}
		err := rows.Scan(
			&pattern.ID, &pattern.Name, &pattern.Regex, &pattern.MediaTypeID,
			&pattern.Priority, &pattern.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		patterns = append(patterns, pattern)
	}
	return patterns, rows.Err()
}

func (r *PatternRepositoryImpl) GetByMediaType(ctx context.Context, mediaTypeID int) ([]*entity.Pattern, error) {
	query := `
		SELECT id, name, regex, media_type_id, priority, created_at
		FROM patterns
		WHERE media_type_id = $1
		ORDER BY priority DESC, id
	`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var patterns []*entity.Pattern
	for rows.Next() {
		pattern := &entity.Pattern{}
		err := rows.Scan(
			&pattern.ID, &pattern.Name, &pattern.Regex, &pattern.MediaTypeID,
			&pattern.Priority, &pattern.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		patterns = append(patterns, pattern)
	}
	return patterns, rows.Err()
}

func (r *PatternRepositoryImpl) Update(ctx context.Context, pattern *entity.Pattern) error {
	query := `
		UPDATE patterns
		SET name = $1, regex = $2, media_type_id = $3, priority = $4
		WHERE id = $5
	`
	_, err := r.db.Pool.Exec(ctx, query, pattern.Name, pattern.Regex, pattern.MediaTypeID, pattern.Priority, pattern.ID)
	return err
}

func (r *PatternRepositoryImpl) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM patterns WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id)
	return err
}
