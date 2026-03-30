-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS media_type_extensions (
    id SERIAL PRIMARY KEY,
    media_type_id INT NOT NULL REFERENCES media_types(id) ON DELETE CASCADE,
    extension VARCHAR(10) NOT NULL,
    UNIQUE(media_type_id, extension)
);

CREATE INDEX idx_media_type_extensions_media_type_id ON media_type_extensions(media_type_id);

INSERT INTO media_type_extensions (media_type_id, extension) VALUES
(1, '.jpg'), (1, '.jpeg'), (1, '.png'), (1, '.gif'), (1, '.webp'),
(1, '.svg'), (1, '.bmp'), (1, '.ico'), (1, '.avif'), (1, '.heic'),
(2, '.mp4'), (2, '.webm'), (2, '.avi'), (2, '.mov'), (2, '.mkv'),
(2, '.wmv'), (2, '.flv'), (2, '.m4v'),
(2, '.mp3'), (2, '.wav'), (2, '.ogg'), (2, '.flac'), (2, '.aac'),
(2, '.m4a'), (2, '.opus'), (2, '.weba'),
(3, '.pdf'), (3, '.doc'), (3, '.docx'), (3, '.txt'), (3, '.rtf'),
(3, '.odt'), (3, '.xls'), (3, '.xlsx'),
(4, '.zip'), (4, '.rar'), (4, '.7z'), (4, '.tar'), (4, '.gz'),
(4, '.bz2'), (4, '.xz'),
(5, '.bin'), (5, '.exe'), (5, '.apk'), (5, '.dmg');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS media_type_extensions;

-- +goose StatementEnd
