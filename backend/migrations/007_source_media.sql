-- +goose Up
-- +goose StatementBegin

CREATE TABLE source_media (
    id           SERIAL PRIMARY KEY,
    source_id    INT NOT NULL REFERENCES sources(id),
    media_id     UUID NOT NULL REFERENCES media(id) ON DELETE CASCADE,
    request_id   UUID REFERENCES requests(id) ON DELETE SET NULL,
    found_at     TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(source_id, media_id)
);

CREATE INDEX idx_source_media_source ON source_media(source_id);
CREATE INDEX idx_source_media_media ON source_media(media_id);
CREATE INDEX idx_source_media_request ON source_media(request_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS source_media;
-- +goose StatementEnd
