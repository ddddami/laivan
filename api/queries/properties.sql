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

-- name: CreatePropertyUnitType :one
INSERT INTO property_unit_types (property_id, category, name, description, bedroom_count, has_parlour, bathroom_type, kitchen_type)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, property_id, name, description, created_at, updated_at, category, bedroom_count, has_parlour, bathroom_type, kitchen_type;

-- name: ListPropertyUnitTypesByProperty :many
SELECT id, property_id, name, description, created_at, updated_at, category, bedroom_count, has_parlour, bathroom_type, kitchen_type
FROM property_unit_types
WHERE property_id = $1
ORDER BY created_at ASC, id ASC;

-- name: GetPropertyUnitType :one
SELECT id, property_id, name, description, created_at, updated_at, category, bedroom_count, has_parlour, bathroom_type, kitchen_type
FROM property_unit_types
WHERE id = $1;

-- name: CreateAgentOffer :one
INSERT INTO agent_offers (property_unit_type_id, agent_id, title, description, price_kobo, status)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, property_unit_type_id, agent_id, title, description, price_kobo, status, created_at, updated_at;

-- name: ListAgentOffersByPropertyUnitType :many
SELECT id, property_unit_type_id, agent_id, title, description, price_kobo, status, created_at, updated_at
FROM agent_offers
WHERE property_unit_type_id = $1
ORDER BY created_at DESC, id DESC;

-- name: ListPropertiesWithSummary :many
SELECT
  p.id, p.campus_id, p.name, p.area, p.landmark, p.description, p.created_at, p.updated_at,
  COALESCE((SELECT COUNT(*) FROM property_unit_types WHERE property_id = p.id), 0)::integer AS unit_type_count,
  COALESCE((SELECT COUNT(*) FROM agent_offers ao JOIN property_unit_types put ON ao.property_unit_type_id = put.id WHERE put.property_id = p.id AND ao.status = 'available'), 0)::integer AS available_offer_count,
  COALESCE((SELECT MIN(ao.price_kobo) FROM agent_offers ao JOIN property_unit_types put ON ao.property_unit_type_id = put.id WHERE put.property_id = p.id AND ao.status = 'available'), 0)::integer AS lowest_price_kobo,
  COUNT(*) OVER() AS total_count
FROM properties p
WHERE p.campus_id = $1
ORDER BY p.created_at DESC, p.id DESC
LIMIT $2;
