-- name: GetScanByID :one
SELECT * FROM scans
WHERE id = $1 LIMIT 1;

-- name: ListRepositoryScans :many
SELECT * FROM scans
WHERE repository_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CreateScan :one
INSERT INTO scans (
  repository_id, status, created_by
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: UpdateScanStatus :one
UPDATE scans
SET 
  status = $2,
  started_at = CASE WHEN $2 = 'in_progress' AND started_at IS NULL THEN NOW() ELSE started_at END,
  completed_at = CASE WHEN $2 IN ('completed', 'failed') THEN NOW() ELSE completed_at END,
  error_message = $3
WHERE id = $1
RETURNING *;

-- name: GetLatestScanByRepositoryID :one
SELECT * FROM scans
WHERE repository_id = $1
ORDER BY created_at DESC
LIMIT 1; 