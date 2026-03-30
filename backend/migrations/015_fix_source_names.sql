-- +goose Up
-- +goose StatementBegin

UPDATE sources
SET name = SPLIT_PART(SPLIT_PART(base_url, '://', 2), '/', 1)
WHERE name IS NULL OR name = '' OR name = base_url;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- No down migration (data transformation)

-- +goose StatementEnd
