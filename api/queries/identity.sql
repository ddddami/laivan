-- name: GetUserByIdentity :one
SELECT u.*
FROM users u
JOIN user_identities ui ON ui.user_id = u.id
WHERE ui.provider = $1
  AND ui.subject = $2;

-- name: CreateUser :one
INSERT INTO users (email, display_name)
VALUES ($1, $2)
RETURNING *;

-- name: UpdateUserProfile :one
UPDATE users
SET email = $2,
    display_name = $3,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: CreateUserIdentity :exec
INSERT INTO user_identities (user_id, provider, subject)
VALUES ($1, $2, $3);

-- name: CreateSession :one
INSERT INTO sessions (user_id, token_hash, csrf_token_hash, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetActiveSession :one
WITH active_session AS (
    UPDATE sessions
    SET last_used_at = now()
    WHERE token_hash = $1
      AND revoked_at IS NULL
      AND expires_at > now()
    RETURNING *
)
SELECT
    s.id AS session_id,
    s.user_id AS session_user_id,
    s.token_hash,
    s.csrf_token_hash,
    s.created_at AS session_created_at,
    s.expires_at,
    s.revoked_at,
    s.last_used_at,
    u.id AS user_id,
    u.email,
    u.display_name,
    u.status,
    u.created_at AS user_created_at,
    u.updated_at AS user_updated_at
FROM active_session s
JOIN users u ON u.id = s.user_id
WHERE u.status = 'active';

-- name: UpdateSessionCSRFToken :exec
UPDATE sessions
SET csrf_token_hash = $2
WHERE token_hash = $1
  AND revoked_at IS NULL
  AND expires_at > now();

-- name: RevokeSession :one
UPDATE sessions
SET revoked_at = now()
WHERE token_hash = $1
  AND revoked_at IS NULL
RETURNING id, user_id;
