-- +goose Up
-- +goose StatementBegin

-- Переименовать video -> video_audio (Видео/Аудио)
UPDATE media_types SET code = 'video_audio', name = 'Видео/Аудио' WHERE code = 'video';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Вернуть старое значение
UPDATE media_types SET code = 'video', name = 'Видео' WHERE code = 'video_audio';

-- +goose StatementEnd
