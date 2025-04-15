-- +goose Up
-- SQL in this section is executed when the migration is applied
ALTER TABLE scans ADD COLUMN IF NOT EXISTS results_available BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back
ALTER TABLE scans DROP COLUMN IF EXISTS results_available; 