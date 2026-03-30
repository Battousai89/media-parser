-- +goose Up
-- +goose StatementBegin

CREATE TABLE requests (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status_id       INT NOT NULL REFERENCES request_statuses(id) DEFAULT 1,
    media_type_id   INT REFERENCES media_types(id),
    limit_count     INT NOT NULL DEFAULT 10,
    offset_count    INT NOT NULL DEFAULT 0,
    priority        INT NOT NULL DEFAULT 0,
    retry_count     INT NOT NULL DEFAULT 0,
    max_retries     INT NOT NULL DEFAULT 3,
    error_message   TEXT,
    started_at      TIMESTAMP WITH TIME ZONE,
    completed_at    TIMESTAMP WITH TIME ZONE,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_requests_status_created ON requests(status_id, created_at);
CREATE INDEX idx_requests_status ON requests(status_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS requests;
-- +goose StatementEnd
