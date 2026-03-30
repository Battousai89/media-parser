-- +goose Up
-- +goose StatementBegin

ALTER TABLE request_sources 
ADD COLUMN IF NOT EXISTS retry_count INT NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS max_retries INT NOT NULL DEFAULT 3;

CREATE INDEX IF NOT EXISTS idx_request_sources_retry ON request_sources(retry_count);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE request_sources 
DROP COLUMN IF EXISTS retry_count,
DROP COLUMN IF EXISTS max_retries;

-- +goose StatementEnd
