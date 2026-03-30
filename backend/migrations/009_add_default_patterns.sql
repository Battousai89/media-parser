-- +goose Up
-- +goose StatementBegin

-- Добавляем паттерны для parsing изображений
INSERT INTO patterns (name, regex, media_type_id, priority) VALUES
    ('Image - JPG direct', 'https?://[^\s"''<>]+\.jpg(?:\?[^\s"''<>]*)?', 1, 100),
    ('Image - JPEG direct', 'https?://[^\s"''<>]+\.jpeg(?:\?[^\s"''<>]*)?', 1, 100),
    ('Image - PNG direct', 'https?://[^\s"''<>]+\.png(?:\?[^\s"''<>]*)?', 1, 100),
    ('Image - GIF direct', 'https?://[^\s"''<>]+\.gif(?:\?[^\s"''<>]*)?', 1, 100),
    ('Image - WEBP direct', 'https?://[^\s"''<>]+\.webp(?:\?[^\s"''<>]*)?', 1, 100),
    ('Image - SVG direct', 'https?://[^\s"''<>]+\.svg(?:\?[^\s"''<>]*)?', 1, 100),
    ('Image - generic URL', 'https?://[^\s"''<>]*(?:/[^/\s"''<>]+)+\.(?:jpg|jpeg|png|gif|webp|svg|bmp|ico)(?:\?[^\s"''<>]*)?', 1, 50);

-- Добавляем паттерны для parsing видео
INSERT INTO patterns (name, regex, media_type_id, priority) VALUES
    ('Video - MP4 direct', 'https?://[^\s"''<>]+\.mp4(?:\?[^\s"''<>]*)?', 2, 100),
    ('Video - WebM direct', 'https?://[^\s"''<>]+\.webm(?:\?[^\s"''<>]*)?', 2, 100),
    ('Video - AVI direct', 'https?://[^\s"''<>]+\.avi(?:\?[^\s"''<>]*)?', 2, 100),
    ('Video - MOV direct', 'https?://[^\s"''<>]+\.mov(?:\?[^\s"''<>]*)?', 2, 100),
    ('Video - MKV direct', 'https?://[^\s"''<>]+\.mkv(?:\?[^\s"''<>]*)?', 2, 100),
    ('Video - generic URL', 'https?://[^\s"''<>]*(?:/[^/\s"''<>]+)+\.(?:mp4|webm|avi|mov|mkv|flv|wmv)(?:\?[^\s"''<>]*)?', 2, 50);

-- Добавляем паттерны для parsing аудио
INSERT INTO patterns (name, regex, media_type_id, priority) VALUES
    ('Audio - MP3 direct', 'https?://[^\s"''<>]+\.mp3(?:\?[^\s"''<>]*)?', 3, 100),
    ('Audio - WAV direct', 'https?://[^\s"''<>]+\.wav(?:\?[^\s"''<>]*)?', 3, 100),
    ('Audio - OGG direct', 'https?://[^\s"''<>]+\.ogg(?:\?[^\s"''<>]*)?', 3, 100),
    ('Audio - FLAC direct', 'https?://[^\s"''<>]+\.flac(?:\?[^\s"''<>]*)?', 3, 100),
    ('Audio - M4A direct', 'https?://[^\s"''<>]+\.m4a(?:\?[^\s"''<>]*)?', 3, 100),
    ('Audio - generic URL', 'https?://[^\s"''<>]*(?:/[^/\s"''<>]+)+\.(?:mp3|wav|ogg|flac|m4a|aac|wma)(?:\?[^\s"''<>]*)?', 3, 50);

-- Добавляем паттерны для parsing документов
INSERT INTO patterns (name, regex, media_type_id, priority) VALUES
    ('Document - PDF direct', 'https?://[^\s"''<>]+\.pdf(?:\?[^\s"''<>]*)?', 4, 100),
    ('Document - DOCX direct', 'https?://[^\s"''<>]+\.docx(?:\?[^\s"''<>]*)?', 4, 100),
    ('Document - DOC direct', 'https?://[^\s"''<>]+\.doc(?:\?[^\s"''<>]*)?', 4, 100),
    ('Document - XLSX direct', 'https?://[^\s"''<>]+\.xlsx(?:\?[^\s"''<>]*)?', 4, 100),
    ('Document - XLS direct', 'https?://[^\s"''<>]+\.xls(?:\?[^\s"''<>]*)?', 4, 100),
    ('Document - TXT direct', 'https?://[^\s"''<>]+\.txt(?:\?[^\s"''<>]*)?', 4, 100),
    ('Document - generic URL', 'https?://[^\s"''<>]*(?:/[^/\s"''<>]+)+\.(?:pdf|docx|doc|xlsx|xls|txt|rtf|odt|ods)(?:\?[^\s"''<>]*)?', 4, 50);

-- Добавляем паттерны для parsing архивов
INSERT INTO patterns (name, regex, media_type_id, priority) VALUES
    ('Archive - ZIP direct', 'https?://[^\s"''<>]+\.zip(?:\?[^\s"''<>]*)?', 5, 100),
    ('Archive - RAR direct', 'https?://[^\s"''<>]+\.rar(?:\?[^\s"''<>]*)?', 5, 100),
    ('Archive - 7Z direct', 'https?://[^\s"''<>]+\.7z(?:\?[^\s"''<>]*)?', 5, 100),
    ('Archive - TAR direct', 'https?://[^\s"''<>]+\.tar(?:\?[^\s"''<>]*)?', 5, 100),
    ('Archive - GZ direct', 'https?://[^\s"''<>]+\.gz(?:\?[^\s"''<>]*)?', 5, 100),
    ('Archive - generic URL', 'https?://[^\s"''<>]*(?:/[^/\s"''<>]+)+\.(?:zip|rar|7z|tar|gz|bz2)(?:\?[^\s"''<>]*)?', 5, 50);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM patterns WHERE name IN (
    'Image - JPG direct', 'Image - JPEG direct', 'Image - PNG direct', 'Image - GIF direct',
    'Image - WEBP direct', 'Image - SVG direct', 'Image - generic URL',
    'Video - MP4 direct', 'Video - WebM direct', 'Video - AVI direct', 'Video - MOV direct',
    'Video - MKV direct', 'Video - generic URL',
    'Audio - MP3 direct', 'Audio - WAV direct', 'Audio - OGG direct', 'Audio - FLAC direct',
    'Audio - M4A direct', 'Audio - generic URL',
    'Document - PDF direct', 'Document - DOCX direct', 'Document - DOC direct',
    'Document - XLSX direct', 'Document - XLS direct', 'Document - TXT direct', 'Document - generic URL',
    'Archive - ZIP direct', 'Archive - RAR direct', 'Archive - 7Z direct', 'Archive - TAR direct',
    'Archive - GZ direct', 'Archive - generic URL'
);
-- +goose StatementEnd
