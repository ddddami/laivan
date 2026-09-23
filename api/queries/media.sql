-- name: CreateMedia :one
INSERT INTO media (property_id, property_unit_type_id, agent_offer_id, uploaded_by_agent_id, url, object_key, kind, caption, content_type, size_bytes)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, property_id, property_unit_type_id, agent_offer_id, uploaded_by_agent_id, url, object_key, kind, caption, content_type, size_bytes, created_at, removed_at, removed_by_user_id;

-- name: ListMediaByProperty :many
SELECT id, property_id, property_unit_type_id, agent_offer_id, uploaded_by_agent_id, url, object_key, kind, caption, content_type, size_bytes, created_at, removed_at, removed_by_user_id
FROM media
WHERE property_id = $1 AND removed_at IS NULL
ORDER BY created_at ASC, id ASC;

-- name: ListMediaByPropertyUnitType :many
SELECT id, property_id, property_unit_type_id, agent_offer_id, uploaded_by_agent_id, url, object_key, kind, caption, content_type, size_bytes, created_at, removed_at, removed_by_user_id
FROM media
WHERE property_unit_type_id = $1 AND removed_at IS NULL
ORDER BY created_at ASC, id ASC;

-- name: ListMediaByPropertyUnitTypeIDs :many
SELECT id, property_id, property_unit_type_id, agent_offer_id, uploaded_by_agent_id, url, object_key, kind, caption, content_type, size_bytes, created_at, removed_at, removed_by_user_id
FROM media
WHERE property_unit_type_id = ANY($1::uuid[]) AND removed_at IS NULL
ORDER BY property_unit_type_id, created_at ASC, id ASC;

-- name: ListMediaByAgentOffer :many
SELECT id, property_id, property_unit_type_id, agent_offer_id, uploaded_by_agent_id, url, object_key, kind, caption, content_type, size_bytes, created_at, removed_at, removed_by_user_id
FROM media
WHERE agent_offer_id = $1 AND removed_at IS NULL
ORDER BY created_at ASC, id ASC;

-- name: ListMediaByAgentOfferIDs :many
SELECT id, property_id, property_unit_type_id, agent_offer_id, uploaded_by_agent_id, url, object_key, kind, caption, content_type, size_bytes, created_at, removed_at, removed_by_user_id
FROM media
WHERE agent_offer_id = ANY($1::uuid[]) AND removed_at IS NULL
ORDER BY agent_offer_id, created_at ASC, id ASC;

-- name: GetMediaForRemoval :one
SELECT
    m.id,
    m.object_key,
    m.property_id,
    m.property_unit_type_id,
    m.agent_offer_id,
    CASE
        WHEN m.property_id IS NOT NULL THEN 'property'
        WHEN m.property_unit_type_id IS NOT NULL THEN 'property_unit_type'
        ELSE 'agent_offer'
    END AS target_type,
    COALESCE(property_target.campus_id, unit_property.campus_id, offer_property.campus_id) AS campus_id,
    offer_target.agent_id
FROM media m
LEFT JOIN properties property_target ON property_target.id = m.property_id
LEFT JOIN property_unit_types unit_target ON unit_target.id = m.property_unit_type_id
LEFT JOIN properties unit_property ON unit_property.id = unit_target.property_id
LEFT JOIN agent_offers offer_target ON offer_target.id = m.agent_offer_id
LEFT JOIN property_unit_types offer_unit ON offer_unit.id = offer_target.property_unit_type_id
LEFT JOIN properties offer_property ON offer_property.id = offer_unit.property_id
WHERE m.id = $1 AND m.removed_at IS NULL;

-- name: RemoveMedia :one
WITH target AS (
    SELECT m.id,
           COALESCE(property_target.campus_id, unit_property.campus_id, offer_property.campus_id) AS campus_id,
           offer_target.agent_id
    FROM media m
    LEFT JOIN properties property_target ON property_target.id = m.property_id
    LEFT JOIN property_unit_types unit_target ON unit_target.id = m.property_unit_type_id
    LEFT JOIN properties unit_property ON unit_property.id = unit_target.property_id
    LEFT JOIN agent_offers offer_target ON offer_target.id = m.agent_offer_id
    LEFT JOIN property_unit_types offer_unit ON offer_unit.id = offer_target.property_unit_type_id
    LEFT JOIN properties offer_property ON offer_property.id = offer_unit.property_id
    WHERE m.id = $1 AND m.removed_at IS NULL
)
UPDATE media AS m
SET removed_at = now(),
    removed_by_user_id = $2
FROM target
WHERE m.id = target.id
  AND m.removed_at IS NULL
  AND (
      EXISTS (SELECT 1 FROM global_admin_roles WHERE user_id = $2)
      OR EXISTS (SELECT 1 FROM campus_operators WHERE user_id = $2 AND campus_id = target.campus_id)
      OR EXISTS (
          SELECT 1
          FROM agents a
          JOIN agent_campuses ac ON ac.agent_id = a.id
          WHERE a.id = target.agent_id
            AND a.user_id = $2
            AND a.status = 'active'
            AND ac.campus_id = target.campus_id
      )
  )
RETURNING m.id;
