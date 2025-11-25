-- +goose Up
-- +goose StatementBegin
-- Таблица заявок (автоматическая при создании Violation и заявки на закрытие)
CREATE TABLE IF NOT EXISTS violation_requests (
    id UUID PRIMARY KEY,
    violation_id UUID NOT NULL,
    status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'partially_closed', 'closed')),
    created_by_user_id BIGINT NOT NULL,
    comment TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_requests_violation_id ON violation_requests (violation_id);
CREATE INDEX IF NOT EXISTS idx_requests_status ON violation_requests (status);
CREATE INDEX IF NOT EXISTS idx_requests_created_by ON violation_requests (created_by_user_id);
CREATE INDEX IF NOT EXISTS idx_requests_created_at ON violation_requests (created_at);

-- Таблица фото для заявок (единственное место хранения фото)
CREATE TABLE IF NOT EXISTS violation_request_photos (
    id UUID PRIMARY KEY,
    request_id UUID NOT NULL,
    url TEXT NOT NULL,
    thumb_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_request_photos_request_id ON violation_request_photos (request_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS violation_request_photos;
DROP TABLE IF EXISTS violation_requests;
-- +goose StatementEnd

