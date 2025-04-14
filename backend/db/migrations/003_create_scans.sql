-- +goose Up
-- SQL in this section is executed when the migration is applied

-- Scans table
CREATE TABLE scans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    repository_id UUID NOT NULL REFERENCES repositories(id),
    status VARCHAR(50) NOT NULL, -- 'pending', 'in_progress', 'completed', 'failed'
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    error_message TEXT,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_scans_repository_id ON scans(repository_id);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back
DROP INDEX IF EXISTS idx_scans_repository_id;
DROP TABLE IF EXISTS scans; 