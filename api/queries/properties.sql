-- name: CreateProperty :one
INSERT INTO properties (campus_id, name, area, landmark, description)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, campus_id, name, area, landmark, description, created_at, updated_at, version;

-- name: GetProperty :one
SELECT id, campus_id, name, area, landmark, description, created_at, updated_at, version
FROM properties
WHERE id = $1;

-- name: GetPropertyUnitType :one
SELECT id, property_id, name, description, created_at, updated_at, category, bedroom_count, has_parlour, bathroom_type, kitchen_type, notes, version
FROM property_unit_types
WHERE id = $1;

-- name: UpdatePropertyUnitType :one
UPDATE property_unit_types
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
WHERE id = sqlc.arg('id')
  AND version = sqlc.arg('expected_version')
RETURNING id, property_id, name, description, created_at, updated_at, category, bedroom_count, has_parlour, bathroom_type, kitchen_type, notes, version;

-- name: UpdateProperty :one
UPDATE properties
SET name = COALESCE(sqlc.narg('name'), name),
    area = COALESCE(sqlc.narg('area'), area),
    landmark = COALESCE(sqlc.narg('landmark'), landmark),
    description = COALESCE(sqlc.narg('description'), description),
    version = version + 1,
    updated_at = now()
WHERE id = sqlc.arg('id')
  AND version = sqlc.arg('expected_version')
RETURNING id, campus_id, name, area, landmark, description, created_at, updated_at, version;

-- name: ListProperties :many
SELECT id, campus_id, name, area, landmark, description, created_at, updated_at
FROM properties
WHERE campus_id = $1
ORDER BY created_at DESC, id DESC
LIMIT $2;

-- name: CreatePropertyUnitType :one
INSERT INTO property_unit_types (property_id, category, name, description, notes, bedroom_count, has_parlour, bathroom_type, kitchen_type)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, property_id, name, description, created_at, updated_at, category, bedroom_count, has_parlour, bathroom_type, kitchen_type, notes, version;

-- name: ListPropertyUnitTypesByProperty :many
SELECT id, property_id, name, description, created_at, updated_at, category, bedroom_count, has_parlour, bathroom_type, kitchen_type, notes, version
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
RETURNING id, property_unit_type_id, agent_id, title, description, price_kobo, status, created_at, updated_at, notes, version;

-- name: ListAgentOffersByPropertyUnitType :many
SELECT id, property_unit_type_id, agent_id, title, description, price_kobo, status, created_at, updated_at, notes, version
FROM agent_offers
WHERE property_unit_type_id = $1
ORDER BY created_at DESC, id DESC;

-- name: GetAgentOffer :one
SELECT id, property_unit_type_id, agent_id, title, description, price_kobo, status, created_at, updated_at, notes, version
FROM agent_offers
WHERE id = $1;

-- name: UpdateAgentOffer :one
UPDATE agent_offers
SET title = COALESCE(sqlc.narg('title'), title),
    description = COALESCE(sqlc.narg('description'), description),
    notes = COALESCE(sqlc.narg('notes'), notes),
    price_kobo = COALESCE(sqlc.narg('price_kobo'), price_kobo),
    status = COALESCE(sqlc.narg('status'), status),
    version = version + 1,
    updated_at = now()
WHERE id = sqlc.arg('id')
  AND version = sqlc.arg('expected_version')
RETURNING id, property_unit_type_id, agent_id, title, description, price_kobo, status, created_at, updated_at, notes, version;

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
  a.display_name AS agent_display_name
FROM agent_offers ao
JOIN agents a ON a.id = ao.agent_id
WHERE ao.property_unit_type_id = ANY($1::uuid[])
ORDER BY ao.property_unit_type_id, ao.created_at DESC, ao.id DESC;

-- name: ListPropertiesWithSummary :many
SELECT
  p.id, p.campus_id, p.name, p.area, p.landmark, p.description, p.created_at, p.updated_at, p.version,
  COALESCE((SELECT COUNT(*) FROM property_unit_types WHERE property_id = p.id), 0)::integer AS unit_type_count,
  COALESCE((SELECT COUNT(*) FROM agent_offers ao JOIN property_unit_types put ON ao.property_unit_type_id = put.id WHERE put.property_id = p.id AND ao.status = 'available'), 0)::integer AS available_offer_count,
  COALESCE((SELECT MIN(ao.price_kobo) FROM agent_offers ao JOIN property_unit_types put ON ao.property_unit_type_id = put.id WHERE put.property_id = p.id AND ao.status = 'available'), 0)::integer AS lowest_price_kobo,
  COUNT(*) OVER() AS total_count
FROM properties p
WHERE p.campus_id = $1
ORDER BY p.created_at DESC, p.id DESC
LIMIT $2;
