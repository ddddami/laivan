-- name: CreateProperty :one
INSERT INTO properties (campus_id, name, area, landmark, description)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, campus_id, name, area, landmark, description, created_at, updated_at;

-- name: GetProperty :one
SELECT id, campus_id, name, area, landmark, description, created_at, updated_at
FROM properties
WHERE id = $1;
