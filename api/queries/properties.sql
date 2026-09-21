-- name: CreateProperty :one
INSERT INTO properties (campus_id, name, area, landmark, description, created_by_user_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, campus_id, name, area, landmark, description, created_at, updated_at, version, created_by_user_id;

-- name: GetProperty :one
SELECT id, campus_id, name, area, landmark, description, created_at, updated_at, version, created_by_user_id
FROM properties
WHERE id = $1;

-- name: GetPropertyUnitType :one
SELECT id, property_id, name, description, created_at, updated_at, category, bedroom_count, has_parlour, bathroom_type, kitchen_type, notes, version, created_by_user_id
FROM property_unit_types
WHERE id = $1;

-- name: UpdatePropertyUnitType :one
UPDATE property_unit_types AS put_target
SET category = COALESCE(sqlc.narg('category'), category),
    name = COALESCE(sqlc.narg('name'), name),
    description = COALESCE(sqlc.narg('description'), description),
    notes = COALESCE(sqlc.narg('notes'), notes),
    bedroom_count = COALESCE(sqlc.narg('bedroom_count'), bedroom_count),
    has_parlour = COALESCE(sqlc.narg('has_parlour'), has_parlour),
    bathroom_type = COALESCE(sqlc.narg('bathroom_type'), bathroom_type),
    kitchen_type = COALESCE(sqlc.narg('kitchen_type'), kitchen_type),
    version = version + 1,
    updated_at = now()
WHERE put_target.id = sqlc.arg('id')
  AND put_target.version = sqlc.arg('expected_version')
  AND (
    EXISTS (
      SELECT 1
      FROM global_admin_roles gar
      WHERE gar.user_id = sqlc.arg('actor_user_id')
    )
    OR EXISTS (
      SELECT 1
      FROM property_unit_types put
      JOIN properties p ON p.id = put.property_id
      JOIN campus_operators co ON co.campus_id = p.campus_id
      WHERE put.id = put_target.id
        AND co.user_id = sqlc.arg('actor_user_id')
    )
  )
RETURNING id, property_id, name, description, created_at, updated_at, category, bedroom_count, has_parlour, bathroom_type, kitchen_type, notes, version, created_by_user_id;

-- name: UpdateProperty :one
UPDATE properties AS p
SET name = COALESCE(sqlc.narg('name'), name),
    area = COALESCE(sqlc.narg('area'), area),
    landmark = COALESCE(sqlc.narg('landmark'), landmark),
    description = COALESCE(sqlc.narg('description'), description),
    version = version + 1,
    updated_at = now()
WHERE p.id = sqlc.arg('id')
  AND p.version = sqlc.arg('expected_version')
  AND (
    EXISTS (
      SELECT 1
      FROM global_admin_roles gar
      WHERE gar.user_id = sqlc.arg('actor_user_id')
    )
    OR EXISTS (
      SELECT 1
      FROM campus_operators co
      WHERE co.user_id = sqlc.arg('actor_user_id')
        AND co.campus_id = p.campus_id
    )
  )
RETURNING id, campus_id, name, area, landmark, description, created_at, updated_at, version, created_by_user_id;

-- name: CreatePropertyUnitType :one
INSERT INTO property_unit_types (property_id, category, name, description, notes, bedroom_count, has_parlour, bathroom_type, kitchen_type, created_by_user_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, property_id, name, description, created_at, updated_at, category, bedroom_count, has_parlour, bathroom_type, kitchen_type, notes, version, created_by_user_id;

-- name: ListPropertyUnitTypesByProperty :many
SELECT id, property_id, name, description, created_at, updated_at, category, bedroom_count, has_parlour, bathroom_type, kitchen_type, notes, version, created_by_user_id
FROM property_unit_types
WHERE property_id = $1
ORDER BY created_at ASC, id ASC;

-- name: GetPropertyUnitTypeMediaTarget :one
SELECT put.id AS property_unit_type_id, put.property_id, p.campus_id
FROM property_unit_types put
JOIN properties p ON p.id = put.property_id
WHERE put.id = $1;

-- name: GetAgentOfferMediaTarget :one
SELECT ao.id AS agent_offer_id, ao.property_unit_type_id, put.property_id, p.campus_id, ao.agent_id
FROM agent_offers ao
JOIN property_unit_types put ON put.id = ao.property_unit_type_id
JOIN properties p ON p.id = put.property_id
WHERE ao.id = $1;

-- name: CreateAuthorizedAgentOffer :one
INSERT INTO agent_offers (property_unit_type_id, agent_id, title, description, notes, price_kobo, status)
SELECT $1, $2, $3, $4, $5, $6, $7
FROM property_unit_types put
JOIN properties p ON p.id = put.property_id
JOIN agents a ON a.id = $2 AND a.status = 'active'
JOIN agent_campuses ac ON ac.agent_id = a.id AND ac.campus_id = p.campus_id
WHERE put.id = $1
RETURNING id, property_unit_type_id, agent_id, title, description, price_kobo, status, created_at, updated_at, notes, version, archived_at;

-- name: ListAgentOffersByPropertyUnitType :many
SELECT id, property_unit_type_id, agent_id, title, description, price_kobo, status, created_at, updated_at, notes, version, archived_at
FROM agent_offers
WHERE property_unit_type_id = $1 AND archived_at IS NULL
ORDER BY created_at DESC, id DESC;

-- name: GetAgentOffer :one
SELECT id, property_unit_type_id, agent_id, title, description, price_kobo, status, created_at, updated_at, notes, version, archived_at
FROM agent_offers
WHERE id = $1;

-- name: UpdateAgentOffer :one
UPDATE agent_offers AS ao
SET title = COALESCE(sqlc.narg('title'), title),
    description = COALESCE(sqlc.narg('description'), description),
    notes = COALESCE(sqlc.narg('notes'), notes),
    price_kobo = COALESCE(sqlc.narg('price_kobo'), price_kobo),
    status = COALESCE(sqlc.narg('status'), status),
    version = version + 1,
    updated_at = now()
WHERE ao.id = sqlc.arg('id')
  AND ao.version = sqlc.arg('expected_version')
  AND ao.archived_at IS NULL
  AND (
    EXISTS (
      SELECT 1
      FROM global_admin_roles gar
      WHERE gar.user_id = sqlc.arg('actor_user_id')
    )
    OR EXISTS (
      SELECT 1
      FROM agents a
      JOIN agent_campuses ac ON ac.agent_id = a.id
      JOIN property_unit_types put ON put.id = ao.property_unit_type_id
      JOIN properties p ON p.id = put.property_id
      WHERE a.id = ao.agent_id
        AND a.user_id = sqlc.arg('actor_user_id')
        AND a.status = 'active'
        AND ac.campus_id = p.campus_id
    )
    OR EXISTS (
      SELECT 1
      FROM property_unit_types put
      JOIN properties p ON p.id = put.property_id
      JOIN campus_operators co ON co.campus_id = p.campus_id
      WHERE put.id = ao.property_unit_type_id
        AND co.user_id = sqlc.arg('actor_user_id')
    )
  )
RETURNING id, property_unit_type_id, agent_id, title, description, price_kobo, status, created_at, updated_at, notes, version, archived_at;

-- name: ArchiveAgentOffer :one
UPDATE agent_offers AS ao
SET status = 'unavailable',
    archived_at = now(),
    version = version + 1,
    updated_at = now()
WHERE ao.id = sqlc.arg('id')
  AND ao.version = sqlc.arg('expected_version')
  AND ao.archived_at IS NULL
  AND (
    EXISTS (
      SELECT 1
      FROM global_admin_roles gar
      WHERE gar.user_id = sqlc.arg('actor_user_id')
    )
    OR EXISTS (
      SELECT 1
      FROM agents a
      JOIN agent_campuses ac ON ac.agent_id = a.id
      JOIN property_unit_types put ON put.id = ao.property_unit_type_id
      JOIN properties p ON p.id = put.property_id
      WHERE a.id = ao.agent_id
        AND a.user_id = sqlc.arg('actor_user_id')
        AND a.status = 'active'
        AND ac.campus_id = p.campus_id
    )
    OR EXISTS (
      SELECT 1
      FROM property_unit_types put
      JOIN properties p ON p.id = put.property_id
      JOIN campus_operators co ON co.campus_id = p.campus_id
      WHERE put.id = ao.property_unit_type_id
        AND co.user_id = sqlc.arg('actor_user_id')
    )
  )
RETURNING id, property_unit_type_id, agent_id, title, description, price_kobo, status, created_at, updated_at, notes, version, archived_at;

-- name: GetPropertyMutationStatus :one
SELECT
  p.version,
  (
    EXISTS (
      SELECT 1
      FROM global_admin_roles gar
      WHERE gar.user_id = sqlc.arg('actor_user_id')
    )
    OR EXISTS (
      SELECT 1
      FROM campus_operators co
      WHERE co.user_id = sqlc.arg('actor_user_id')
        AND co.campus_id = p.campus_id
    )
  ) AS is_authorized
FROM properties p
WHERE p.id = sqlc.arg('id');

-- name: GetPropertyUnitTypeMutationStatus :one
SELECT
  put.version,
  (
    EXISTS (
      SELECT 1
      FROM global_admin_roles gar
      WHERE gar.user_id = sqlc.arg('actor_user_id')
    )
    OR EXISTS (
      SELECT 1
      FROM properties p
      JOIN campus_operators co ON co.campus_id = p.campus_id
      WHERE p.id = put.property_id
        AND co.user_id = sqlc.arg('actor_user_id')
    )
  ) AS is_authorized
FROM property_unit_types put
WHERE put.id = sqlc.arg('id');

-- name: GetAgentOfferMutationStatus :one
SELECT
  ao.version,
  ao.archived_at,
  (
    EXISTS (
      SELECT 1
      FROM global_admin_roles gar
      WHERE gar.user_id = sqlc.arg('actor_user_id')
    )
    OR EXISTS (
      SELECT 1
      FROM agents a
      JOIN agent_campuses ac ON ac.agent_id = a.id
      JOIN property_unit_types put ON put.id = ao.property_unit_type_id
      JOIN properties p ON p.id = put.property_id
      WHERE a.id = ao.agent_id
        AND a.user_id = sqlc.arg('actor_user_id')
        AND a.status = 'active'
        AND ac.campus_id = p.campus_id
    )
    OR EXISTS (
      SELECT 1
      FROM property_unit_types put
      JOIN properties p ON p.id = put.property_id
      JOIN campus_operators co ON co.campus_id = p.campus_id
      WHERE put.id = ao.property_unit_type_id
        AND co.user_id = sqlc.arg('actor_user_id')
    )
  ) AS is_authorized
FROM agent_offers ao
WHERE ao.id = sqlc.arg('id');

-- name: ListAgentOfferDetailsByPropertyUnitTypeIDs :many
SELECT
  ao.id,
  ao.property_unit_type_id,
  ao.agent_id,
  ao.title,
  ao.description,
  ao.price_kobo,
  ao.status,
  ao.created_at,
  ao.updated_at,
  ao.notes,
  ao.version,
  ao.archived_at,
  a.display_name AS agent_display_name
FROM agent_offers ao
JOIN agents a ON a.id = ao.agent_id
WHERE ao.property_unit_type_id = ANY($1::uuid[]) AND ao.archived_at IS NULL
ORDER BY ao.property_unit_type_id, ao.created_at DESC, ao.id DESC;

-- name: ListPropertiesWithSummary :many
SELECT
  p.id, p.campus_id, p.name, p.area, p.landmark, p.description, p.created_at, p.updated_at, p.version,
  COALESCE((SELECT COUNT(*) FROM property_unit_types WHERE property_id = p.id), 0)::integer AS unit_type_count,
  COALESCE((SELECT COUNT(*) FROM agent_offers ao JOIN property_unit_types put ON ao.property_unit_type_id = put.id WHERE put.property_id = p.id AND ao.status = 'available' AND ao.archived_at IS NULL), 0)::integer AS available_offer_count,
  COALESCE((SELECT MIN(ao.price_kobo) FROM agent_offers ao JOIN property_unit_types put ON ao.property_unit_type_id = put.id WHERE put.property_id = p.id AND ao.status = 'available' AND ao.archived_at IS NULL), 0)::integer AS lowest_price_kobo,
  COUNT(*) OVER() AS total_count
FROM properties p
WHERE p.campus_id = $1
ORDER BY p.created_at DESC, p.id DESC
LIMIT $2;
