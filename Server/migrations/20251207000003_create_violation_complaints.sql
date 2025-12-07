-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS violation_complaints (
    id UUID PRIMARY KEY,
    violation_id UUID NOT NULL,
    request_id UUID,
    user_id BIGINT NOT NULL,
    reason TEXT,
    message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_violation_complaints_violation_id
    ON violation_complaints (violation_id);

CREATE INDEX IF NOT EXISTS idx_violation_complaints_user_violation
    ON violation_complaints (user_id, violation_id);

CREATE INDEX IF NOT EXISTS idx_violation_complaints_request_id
    ON violation_complaints (request_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS violation_complaints;
-- +goose StatementEnd


