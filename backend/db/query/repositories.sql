-- name: GetRepositoryByID :one
SELECT * FROM repositories
WHERE id = $1 LIMIT 1;

-- name: GetRepositoryByOwnerAndName :one
SELECT * FROM repositories
WHERE owner = $1 AND name = $2 LIMIT 1;

-- name: ListRepositories :many
SELECT * FROM repositories
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListUserRepositories :many
SELECT * FROM repositories
WHERE created_by = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CreateRepository :one
INSERT INTO repositories (
  owner, name, url, clone_url, description, created_by
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: UpdateRepository :one
UPDATE repositories
SET 
  description = COALESCE($2, description)
WHERE id = $1
RETURNING *;

-- name: DeleteRepository :exec
DELETE FROM repositories
WHERE id = $1; 