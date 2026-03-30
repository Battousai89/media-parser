package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/media-parser/backend/internal/model/entity"
)

type APITokenRepositoryImpl struct {
	db *PostgresDB
}

func NewAPITokenRepository(db *PostgresDB) *APITokenRepositoryImpl {
	return &APITokenRepositoryImpl{db: db}
}

func (r *APITokenRepositoryImpl) Create(ctx context.Context, token *entity.APIToken) error {
	query := `
		INSERT INTO api_tokens (token, name, active, expires_at, permissions, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id
	`
	return r.db.Pool.QueryRow(ctx, query,
		token.Token, token.Name, token.Active, token.ExpiresAt, token.Permissions,
	).Scan(&token.ID)
}

func (r *APITokenRepositoryImpl) GetByToken(ctx context.Context, token string) (*entity.APIToken, error) {
	query := `
		SELECT id, token, name, active, expires_at, permissions, created_at, last_used_at
		FROM api_tokens
		WHERE token = $1
	`
	t := &entity.APIToken{}
	err := r.db.Pool.QueryRow(ctx, query, token).Scan(
		&t.ID, &t.Token, &t.Name, &t.Active, &t.ExpiresAt, &t.Permissions,
		&t.CreatedAt, &t.LastUsedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return t, nil
}

func (r *APITokenRepositoryImpl) GetByID(ctx context.Context, id int) (*entity.APIToken, error) {
	query := `
		SELECT id, token, name, active, expires_at, permissions, created_at, last_used_at
		FROM api_tokens
		WHERE id = $1
	`
	t := &entity.APIToken{}
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.Token, &t.Name, &t.Active, &t.ExpiresAt, &t.Permissions,
		&t.CreatedAt, &t.LastUsedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return t, nil
}

func (r *APITokenRepositoryImpl) UpdateLastUsed(ctx context.Context, id int) error {
	query := `UPDATE api_tokens SET last_used_at = NOW() WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id)
	return err
}

func (r *APITokenRepositoryImpl) Update(ctx context.Context, token *entity.APIToken) error {
	query := `
		UPDATE api_tokens
		SET name = $1, active = $2, expires_at = $3, permissions = $4
		WHERE id = $5
	`
	_, err := r.db.Pool.Exec(ctx, query,
		token.Name, token.Active, token.ExpiresAt, token.Permissions, token.ID,
	)
	return err
}

func (r *APITokenRepositoryImpl) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM api_tokens WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id)
	return err
}
