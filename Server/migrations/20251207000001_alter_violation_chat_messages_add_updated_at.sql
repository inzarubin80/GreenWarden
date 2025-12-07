-- +goose Up
-- +goose StatementBegin
ALTER TABLE violation_chat_messages
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE violation_chat_messages
    DROP COLUMN IF EXISTS updated_at;
-- +goose StatementEnd


