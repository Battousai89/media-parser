-- +goose Up
-- +goose StatementBegin

CREATE TABLE api_tokens (
    id           SERIAL PRIMARY KEY,
    token        VARCHAR(255) NOT NULL UNIQUE,
    name         VARCHAR(255),
    active       BOOLEAN DEFAULT true,
    expires_at   TIMESTAMP WITH TIME ZONE,
    permissions  JSONB,
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_used_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_api_tokens_token ON api_tokens(token);
CREATE INDEX idx_api_tokens_active ON api_tokens(active) WHERE active = true;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS api_tokens;
-- +goose StatementEnd
