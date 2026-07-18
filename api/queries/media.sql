-- name: CreateMedia :one
INSERT INTO media (property_id, property_unit_type_id, agent_offer_id, uploaded_by_agent_id, url, object_key, kind, caption, content_type, size_bytes)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, property_id, property_unit_type_id, agent_offer_id, uploaded_by_agent_id, url, object_key, kind, caption, content_type, size_bytes, created_at;

-- name: ListMediaByProperty :many
SELECT id, property_id, property_unit_type_id, agent_offer_id, uploaded_by_agent_id, url, object_key, kind, caption, content_type, size_bytes, created_at
FROM media
WHERE property_id = $1
ORDER BY created_at ASC, id ASC;

-- name: ListMediaByPropertyUnitType :many
SELECT id, property_id, property_unit_type_id, agent_offer_id, uploaded_by_agent_id, url, object_key, kind, caption, content_type, size_bytes, created_at
FROM media
WHERE property_unit_type_id = $1
ORDER BY created_at ASC, id ASC;

-- name: ListMediaByPropertyUnitTypeIDs :many
SELECT id, property_id, property_unit_type_id, agent_offer_id, uploaded_by_agent_id, url, object_key, kind, caption, content_type, size_bytes, created_at
FROM media
WHERE property_unit_type_id = ANY($1::uuid[])
ORDER BY property_unit_type_id, created_at ASC, id ASC;

-- name: ListMediaByAgentOffer :many
SELECT id, property_id, property_unit_type_id, agent_offer_id, uploaded_by_agent_id, url, object_key, kind, caption, content_type, size_bytes, created_at
FROM media
WHERE agent_offer_id = $1
ORDER BY created_at ASC, id ASC;
