-- +goose Up
-- Миграция: добавление полей для проверки статуса источников

ALTER TABLE sources ADD COLUMN IF NOT EXISTS last_checked_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE sources ADD COLUMN IF NOT EXISTS check_error_message TEXT;

-- Индекс для оптимизации запросов по last_checked_at
CREATE INDEX IF NOT EXISTS idx_sources_last_checked_at ON sources(last_checked_at);

-- +goose Down
ALTER TABLE sources DROP COLUMN IF EXISTS last_checked_at;
ALTER TABLE sources DROP COLUMN IF EXISTS check_error_message;
