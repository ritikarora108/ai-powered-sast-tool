-- +goose Up
-- SQL in this section is executed when the migration is applied

-- Add last_scan_at and status columns to repositories table
ALTER TABLE repositories 
  ADD COLUMN last_scan_at TIMESTAMPTZ,
  ADD COLUMN status VARCHAR(50);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back

-- Remove the columns
ALTER TABLE repositories 
  DROP COLUMN IF EXISTS last_scan_at,
  DROP COLUMN IF EXISTS status; 