-- +goose Up
-- +goose StatementBegin

-- Статусы запросов
CREATE TABLE request_statuses (
    id      SERIAL PRIMARY KEY,
    code    VARCHAR(50) NOT NULL UNIQUE,
    name    VARCHAR(100) NOT NULL
);

INSERT INTO request_statuses (code, name) VALUES
    ('pending', 'Ожидает'),
    ('processing', 'В процессе'),
    ('completed', 'Завершён'),
    ('failed', 'Ошибка'),
    ('partial', 'Частично');

-- Типы медиа
CREATE TABLE media_types (
    id      SERIAL PRIMARY KEY,
    code    VARCHAR(50) NOT NULL UNIQUE,
    name    VARCHAR(100) NOT NULL
);

INSERT INTO media_types (code, name) VALUES
    ('image', 'Изображение'),
    ('video', 'Видео/Аудио'),
    ('document', 'Документ'),
    ('archive', 'Архив'),
    ('other', 'Другое');

-- Статусы источников
CREATE TABLE source_statuses (
    id      SERIAL PRIMARY KEY,
    code    VARCHAR(50) NOT NULL UNIQUE,
    name    VARCHAR(100) NOT NULL
);

INSERT INTO source_statuses (code, name) VALUES
    ('active', 'Активен'),
    ('inactive', 'Неактивен'),
    ('error', 'Ошибка'),
    ('blocked', 'Заблокирован');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS source_statuses;
DROP TABLE IF EXISTS media_types;
DROP TABLE IF EXISTS request_statuses;
-- +goose StatementEnd
