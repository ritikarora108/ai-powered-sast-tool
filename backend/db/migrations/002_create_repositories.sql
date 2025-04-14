-- +goose Up
-- SQL in this section is executed when the migration is applied

-- Repositories table
CREATE TABLE repositories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    url TEXT NOT NULL,
    clone_url TEXT NOT NULL,
    description TEXT,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (owner, name)
);

-- Create indexes
CREATE INDEX idx_repositories_created_by ON repositories(created_by);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back
DROP INDEX IF EXISTS idx_repositories_created_by;
DROP TABLE IF EXISTS repositories; 