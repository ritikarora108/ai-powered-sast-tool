-- +goose Up
-- SQL in this section is executed when the migration is applied

-- Add role and notification preferences to users table
ALTER TABLE users ADD COLUMN role VARCHAR(50) DEFAULT 'user' NOT NULL;
ALTER TABLE users ADD COLUMN receive_notifications BOOLEAN DEFAULT TRUE NOT NULL;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back
ALTER TABLE users DROP COLUMN IF EXISTS receive_notifications;
ALTER TABLE users DROP COLUMN IF EXISTS role; 