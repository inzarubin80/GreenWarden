-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS violation_photos (
    id UUID PRIMARY KEY,
    violation_id UUID NOT NULL,
    url TEXT NOT NULL,
    thumb_url TEXT
);
CREATE INDEX IF NOT EXISTS idx_violation_photos_violation_id ON violation_photos (violation_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS violation_photos;
-- +goose StatementEnd


