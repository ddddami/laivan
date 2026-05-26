-- name: CreateProperty :one
INSERT INTO properties (campus_id, name, area, landmark, description)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, campus_id, name, area, landmark, description, created_at, updated_at;

-- name: GetProperty :one
SELECT id, campus_id, name, area, landmark, description, created_at, updated_at
FROM properties
WHERE id = $1;

-- name: ListProperties :many
SELECT id, campus_id, name, area, landmark, description, created_at, updated_at
FROM properties
WHERE campus_id = $1
ORDER BY created_at DESC, id DESC
LIMIT $2;
