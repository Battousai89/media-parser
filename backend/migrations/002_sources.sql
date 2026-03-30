-- +goose Up
-- +goose StatementBegin

CREATE TABLE sources (
    id           SERIAL PRIMARY KEY,
    name         VARCHAR(255) NOT NULL,
    base_url     VARCHAR(2048) NOT NULL,
    status_id    INT NOT NULL REFERENCES source_statuses(id) DEFAULT 1,
    updated_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_sources_status ON sources(status_id);
CREATE INDEX idx_sources_status_active ON sources(status_id) WHERE status_id = 1;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS sources;
-- +goose StatementEnd
