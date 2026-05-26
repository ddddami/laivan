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

-- name: CreateRoomType :one
INSERT INTO room_types (property_id, name, description)
VALUES ($1, $2, $3)
RETURNING id, property_id, name, description, created_at, updated_at;

-- name: ListRoomTypesByProperty :many
SELECT id, property_id, name, description, created_at, updated_at
FROM room_types
WHERE property_id = $1
ORDER BY created_at ASC, id ASC;

-- name: GetRoomType :one
SELECT id, property_id, name, description, created_at, updated_at
FROM room_types
WHERE id = $1;

-- name: CreateAgentOffer :one
INSERT INTO agent_offers (room_type_id, agent_id, title, description, price_kobo, status)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, room_type_id, agent_id, title, description, price_kobo, status, created_at, updated_at;

-- name: ListAgentOffersByRoomType :many
SELECT id, room_type_id, agent_id, title, description, price_kobo, status, created_at, updated_at
FROM agent_offers
WHERE room_type_id = $1
ORDER BY created_at DESC, id DESC;
