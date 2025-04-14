-- name: CreateRepository :one
INSERT INTO repositories (
  name,
  owner,
  url,
  clone_url,
  is_private
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetRepository :one
SELECT * FROM repositories
WHERE id = $1;

-- name: ListRepositories :many
SELECT * FROM repositories
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateRepository :one
UPDATE repositories
SET
  name = $2,
  owner = $3,
  url = $4,
  clone_url = $5,
  is_private = $6,
  updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteRepository :exec
DELETE FROM repositories
WHERE id = $1; 