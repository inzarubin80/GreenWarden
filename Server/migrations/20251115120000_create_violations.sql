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
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS violations;
-- +goose StatementEnd


