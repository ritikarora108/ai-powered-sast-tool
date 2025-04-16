-- +goose Up
-- SQL in this section is executed when the migration is applied

-- Find the first admin user or any user to set as repository owner
WITH first_user AS (
    SELECT id FROM users ORDER BY created_at LIMIT 1
)
-- Update all repositories with NULL created_by to use first user
UPDATE repositories
SET created_by = (SELECT id FROM first_user)
WHERE created_by IS NULL;

-- Update all scans with NULL created_by to use the repository owner
UPDATE scans s
SET created_by = r.created_by
FROM repositories r
WHERE s.repository_id = r.id AND s.created_by IS NULL;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back
-- No down migration needed for this update 