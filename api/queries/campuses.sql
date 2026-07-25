-- name: GetCampusBySlug :one
SELECT id, slug, name, short_name, is_active, created_at, updated_at
FROM campuses
WHERE slug = $1
  AND is_active = true;
