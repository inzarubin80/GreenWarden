-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS violation_chat_messages (
    id UUID PRIMARY KEY,
    violation_id UUID NOT NULL,
    user_id BIGINT NOT NULL,
    text TEXT NOT NULL,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_violation_chat_messages_violation_id_created_at
    ON violation_chat_messages (violation_id, created_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS violation_chat_messages;
-- +goose StatementEnd


