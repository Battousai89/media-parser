package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/media-parser/backend/internal/model/entity"
)

type DictionaryRepositoryImpl struct {
	db *PostgresDB
}

func NewDictionaryRepository(db *PostgresDB) *DictionaryRepositoryImpl {
	return &DictionaryRepositoryImpl{db: db}
}

func (r *DictionaryRepositoryImpl) GetRequestStatuses(ctx context.Context) ([]*entity.RequestStatus, error) {
	query := `SELECT id, code, name FROM request_statuses ORDER BY id`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statuses []*entity.RequestStatus
	for rows.Next() {
		s := &entity.RequestStatus{}
		err := rows.Scan(&s.ID, &s.Code, &s.Name)
		if err != nil {
			return nil, err
		}
		statuses = append(statuses, s)
	}
	return statuses, rows.Err()
}

func (r *DictionaryRepositoryImpl) GetMediaTypes(ctx context.Context) ([]*entity.MediaType, error) {
	query := `SELECT id, code, name FROM media_types ORDER BY id`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var types []*entity.MediaType
	for rows.Next() {
		t := &entity.MediaType{}
		err := rows.Scan(&t.ID, &t.Code, &t.Name)
		if err != nil {
			return nil, err
		}
		types = append(types, t)
	}
	return types, rows.Err()
}

func (r *DictionaryRepositoryImpl) GetSourceStatuses(ctx context.Context) ([]*entity.SourceStatus, error) {
	query := `SELECT id, code, name FROM source_statuses ORDER BY id`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statuses []*entity.SourceStatus
	for rows.Next() {
		s := &entity.SourceStatus{}
		err := rows.Scan(&s.ID, &s.Code, &s.Name)
		if err != nil {
			return nil, err
		}
		statuses = append(statuses, s)
	}
	return statuses, rows.Err()
}

func (r *DictionaryRepositoryImpl) GetRequestStatusByCode(ctx context.Context, code string) (*entity.RequestStatus, error) {
	query := `SELECT id, code, name FROM request_statuses WHERE code = $1`
	s := &entity.RequestStatus{}
	err := r.db.Pool.QueryRow(ctx, query, code).Scan(&s.ID, &s.Code, &s.Name)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return s, nil
}

func (r *DictionaryRepositoryImpl) GetMediaTypeByCode(ctx context.Context, code string) (*entity.MediaType, error) {
	query := `SELECT id, code, name FROM media_types WHERE code = $1`
	t := &entity.MediaType{}
	err := r.db.Pool.QueryRow(ctx, query, code).Scan(&t.ID, &t.Code, &t.Name)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return t, nil
}

func (r *DictionaryRepositoryImpl) GetSourceStatusByCode(ctx context.Context, code string) (*entity.SourceStatus, error) {
	query := `SELECT id, code, name FROM source_statuses WHERE code = $1`
	s := &entity.SourceStatus{}
	err := r.db.Pool.QueryRow(ctx, query, code).Scan(&s.ID, &s.Code, &s.Name)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return s, nil
}
