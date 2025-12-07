-- +goose Up
-- +goose StatementBegin
-- Обновляем CHECK constraint для статусов Violation (добавляем partially_resolved)
ALTER TABLE violations DROP CONSTRAINT IF EXISTS violations_status_check;
ALTER TABLE violations ADD CONSTRAINT violations_status_check 
    CHECK (status IN ('new','confirmed','resolved','partially_resolved'));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Возвращаем старый CHECK constraint.
-- Сначала нормализуем статусы, чтобы constraint не падал.
ALTER TABLE violations DROP CONSTRAINT IF EXISTS violations_status_check;

-- Любые "новые" статусы приводим к ближайшему допустимому значению.
UPDATE violations
SET status = 'resolved'
WHERE status NOT IN ('new','confirmed','resolved');

ALTER TABLE violations ADD CONSTRAINT violations_status_check
    CHECK (status IN ('new','confirmed','resolved'));
-- +goose StatementEnd

