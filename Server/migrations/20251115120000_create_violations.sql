-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS violations (
    id UUID PRIMARY KEY,
    user_id BIGINT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('garbage','pollution','air','deforestation','other')),
    description TEXT,
    lat DOUBLE PRECISION NOT NULL,
    lng DOUBLE PRECISION NOT NULL,
    status TEXT NOT NULL DEFAULT 'new' CHECK (status IN ('new','confirmed','resolved')),
    confirmations_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_violations_user_id ON violations (user_id);
CREATE INDEX IF NOT EXISTS idx_violations_status ON violations (status);
CREATE INDEX IF NOT EXISTS idx_violations_lng_lat ON violations (lng, lat);
-- child table for photos (no FK per requirements)
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
DROP TABLE IF EXISTS violations;
-- +goose StatementEnd


