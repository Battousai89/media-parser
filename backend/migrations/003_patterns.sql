-- +goose Up
-- +goose StatementBegin

CREATE TABLE patterns (
    id            SERIAL PRIMARY KEY,
    name          VARCHAR(255) NOT NULL,
    regex         TEXT NOT NULL,
    media_type_id INT NOT NULL REFERENCES media_types(id),
    priority      INT DEFAULT 0,
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_patterns_media_type ON patterns(media_type_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS patterns;
-- +goose StatementEnd
