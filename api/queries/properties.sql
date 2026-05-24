-- name: CreateProperty :one
INSERT INTO properties (name, area, landmark, district, description)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, name, area, landmark, district, description, created_at, updated_at;

-- name: GetProperty :one
SELECT id, name, area, landmark, district, description, created_at, updated_at
FROM properties
WHERE id = $1;
