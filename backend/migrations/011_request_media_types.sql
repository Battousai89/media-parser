-- +goose Up
-- Миграция: таблица request_media_types для поддержки нескольких типов медиа в запросе

CREATE TABLE IF NOT EXISTS request_media_types (
    id            SERIAL PRIMARY KEY,
    request_id    UUID NOT NULL REFERENCES requests(id) ON DELETE CASCADE,
    media_type_id INT NOT NULL REFERENCES media_types(id),
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(request_id, media_type_id)
);

CREATE INDEX IF NOT EXISTS idx_request_media_types_request ON request_media_types(request_id);
CREATE INDEX IF NOT EXISTS idx_request_media_types_media_type ON request_media_types(media_type_id);

-- Перенос данных из requests.media_type_id
INSERT INTO request_media_types (request_id, media_type_id)
SELECT id, media_type_id FROM requests WHERE media_type_id IS NOT NULL;

-- Удаляем старый столбец media_type_id из requests
ALTER TABLE requests DROP COLUMN IF EXISTS media_type_id;

-- +goose Down
ALTER TABLE requests ADD COLUMN IF NOT EXISTS media_type_id INT REFERENCES media_types(id);
DROP TABLE IF EXISTS request_media_types;
