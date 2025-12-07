-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS violation_votes (
    id UUID PRIMARY KEY,
    violation_id UUID NOT NULL,
    request_id UUID NOT NULL,
    user_id BIGINT NOT NULL,
    value TEXT NOT NULL CHECK (value IN ('like', 'dislike')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Уникальный голос на заявку: один голос пользователя на каждую violation_request.
CREATE UNIQUE INDEX IF NOT EXISTS idx_violation_votes_request_user
    ON violation_votes (request_id, user_id);
CREATE INDEX IF NOT EXISTS idx_violation_votes_violation_id
    ON violation_votes (violation_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS violation_votes;
-- +goose StatementEnd


