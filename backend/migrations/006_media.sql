-- +goose Up
-- +goose StatementBegin

CREATE TABLE media (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    url           VARCHAR(2048) NOT NULL UNIQUE,
    media_type_id INT NOT NULL REFERENCES media_types(id),
    title         VARCHAR(500),
    description   TEXT,
    file_size     BIGINT,
    mime_type     VARCHAR(255),
    hash          VARCHAR(64),
    meta          JSONB,
    available     BOOLEAN DEFAULT true,
    checked_at    TIMESTAMP WITH TIME ZONE,
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_media_type_created ON media(media_type_id, created_at DESC);
CREATE INDEX idx_media_hash ON media(hash);
CREATE INDEX idx_media_available ON media(available) WHERE available = false;
CREATE INDEX idx_media_url ON media(url);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS media;
-- +goose StatementEnd
