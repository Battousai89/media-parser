-- +goose Up
-- +goose StatementBegin

ALTER TABLE media ADD COLUMN storage_path VARCHAR(500);
CREATE INDEX idx_media_storage_path ON media(storage_path);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_media_storage_path;
ALTER TABLE media DROP COLUMN IF EXISTS storage_path;

-- +goose StatementEnd
