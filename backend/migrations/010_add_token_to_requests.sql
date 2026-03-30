-- +goose Up
-- +goose StatementBegin

ALTER TABLE requests ADD COLUMN IF NOT EXISTS token_id INTEGER REFERENCES api_tokens(id);
CREATE INDEX IF NOT EXISTS idx_requests_token_id ON requests(token_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_requests_token_id;
ALTER TABLE requests DROP COLUMN IF EXISTS token_id;

-- +goose StatementEnd
