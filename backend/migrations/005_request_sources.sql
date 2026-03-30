-- +goose Up
-- +goose StatementBegin

CREATE TABLE request_sources (
    id            SERIAL PRIMARY KEY,
    request_id    UUID NOT NULL REFERENCES requests(id) ON DELETE CASCADE,
    source_id     INT NOT NULL REFERENCES sources(id),
    status_id     INT NOT NULL REFERENCES request_statuses(id) DEFAULT 1,
    media_count   INT NOT NULL DEFAULT 0,
    parsed_count  INT NOT NULL DEFAULT 0,
    error_message TEXT,
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(request_id, source_id)
);

CREATE INDEX idx_request_sources_request ON request_sources(request_id);
CREATE INDEX idx_request_sources_source ON request_sources(source_id);
CREATE INDEX idx_request_sources_status ON request_sources(status_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS request_sources;
-- +goose StatementEnd
