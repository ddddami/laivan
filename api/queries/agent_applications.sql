-- name: IsGlobalAdmin :one
SELECT EXISTS (
    SELECT 1 FROM global_admin_roles WHERE user_id = $1
) AS is_global_admin;

-- name: ListCampusOperatorCampuses :many
SELECT campus_id
FROM campus_operators
WHERE user_id = $1
ORDER BY campus_id;

-- name: GetLinkedAgentAccess :one
SELECT id, status
FROM agents
WHERE user_id = $1;

-- name: ListAgentCampuses :many
SELECT campus_id
FROM agent_campuses
WHERE agent_id = $1
ORDER BY campus_id;

-- name: AssociateAgentWithCampus :exec
INSERT INTO agent_campuses (agent_id, campus_id)
VALUES ($1, $2)
ON CONFLICT (agent_id, campus_id) DO NOTHING;

-- name: LockUserForAgentApplication :one
SELECT id
FROM users
WHERE id = $1
FOR UPDATE;

-- name: CreateAgentApplication :one
INSERT INTO agent_applications (applicant_user_id, campus_id, name, phone_number)
SELECT $1, $2, $3, $4
FROM campuses
WHERE id = $2 AND is_active = true
RETURNING id;

-- name: GetAgentApplication :one
SELECT
    aa.id,
    aa.applicant_user_id,
    aa.campus_id,
    aa.name,
    aa.phone_number,
    aa.status,
    aa.reviewer_user_id,
    aa.agent_id,
    aa.operator_note,
    aa.created_at,
    aa.updated_at,
    aa.decided_at,
    u.email AS applicant_email,
    u.display_name AS applicant_display_name
FROM agent_applications aa
JOIN users u ON u.id = aa.applicant_user_id
WHERE aa.id = $1;

-- name: ListAgentApplicationsByApplicant :many
SELECT
    aa.id,
    aa.applicant_user_id,
    aa.campus_id,
    aa.name,
    aa.phone_number,
    aa.status,
    aa.reviewer_user_id,
    aa.agent_id,
    aa.operator_note,
    aa.created_at,
    aa.updated_at,
    aa.decided_at,
    u.email AS applicant_email,
    u.display_name AS applicant_display_name
FROM agent_applications aa
JOIN users u ON u.id = aa.applicant_user_id
WHERE aa.applicant_user_id = $1
ORDER BY aa.created_at DESC, aa.id DESC;

-- name: ListAgentApplicationsByCampus :many
SELECT
    aa.id,
    aa.applicant_user_id,
    aa.campus_id,
    aa.name,
    aa.phone_number,
    aa.status,
    aa.reviewer_user_id,
    aa.agent_id,
    aa.operator_note,
    aa.created_at,
    aa.updated_at,
    aa.decided_at,
    u.email AS applicant_email,
    u.display_name AS applicant_display_name
FROM agent_applications aa
JOIN users u ON u.id = aa.applicant_user_id
WHERE aa.campus_id = $1
  AND ($2 = '' OR aa.status = $2)
ORDER BY aa.created_at DESC, aa.id DESC;

-- name: CanReviewCampus :one
SELECT EXISTS (
    SELECT 1
    FROM global_admin_roles
    WHERE global_admin_roles.user_id = $1
) OR EXISTS (
    SELECT 1
    FROM campus_operators
    WHERE campus_operators.user_id = $1 AND campus_operators.campus_id = $2
) AS can_review;

-- name: GetAgentApplicationForUpdate :one
SELECT
    aa.id,
    aa.applicant_user_id,
    aa.campus_id,
    aa.name,
    aa.phone_number,
    aa.status,
    aa.reviewer_user_id,
    aa.agent_id,
    aa.operator_note,
    aa.created_at,
    aa.updated_at,
    aa.decided_at,
    u.email AS applicant_email,
    u.display_name AS applicant_display_name
FROM agent_applications aa
JOIN users u ON u.id = aa.applicant_user_id
WHERE aa.id = $1
FOR UPDATE OF aa;

-- name: GetAgent :one
SELECT id, user_id, display_name, phone_number, whatsapp_number, status, created_at, updated_at
FROM agents
WHERE id = $1;

-- name: CreateAgent :one
INSERT INTO agents (user_id, display_name, phone_number, status)
VALUES ($1, $2, $3, 'active')
RETURNING id, user_id, display_name, phone_number, whatsapp_number, status, created_at, updated_at;


-- name: ActivateAgentApplication :one
UPDATE agent_applications
SET status = 'active',
    reviewer_user_id = $2,
    agent_id = $3,
    operator_note = NULLIF($4, ''),
    decided_at = now(),
    updated_at = now()
WHERE id = $1 AND status = 'pending'
RETURNING id;

-- name: DeclineAgentApplication :one
UPDATE agent_applications
SET status = 'declined',
    reviewer_user_id = $2,
    operator_note = NULLIF($3, ''),
    decided_at = now(),
    updated_at = now()
WHERE id = $1 AND status = 'pending'
RETURNING id;

-- name: CreateAuditEvent :exec
INSERT INTO audit_events (actor_user_id, action, resource_type, resource_id, metadata)
VALUES ($1, $2, $3, $4, $5);

-- name: GetAgentForLifecycleUpdate :one
SELECT id, user_id, status
FROM agents
WHERE id = $1
FOR UPDATE;

-- name: CanManageAgent :one
SELECT EXISTS (
    SELECT 1
    FROM global_admin_roles
    WHERE global_admin_roles.user_id = $1
) OR EXISTS (
    SELECT 1
    FROM campus_operators co
    JOIN agent_campuses ac ON ac.campus_id = co.campus_id
    WHERE co.user_id = $1 AND ac.agent_id = $2
) AS can_manage;

-- name: SuspendAgent :one
UPDATE agents
SET status = 'suspended', updated_at = now()
WHERE id = $1 AND status = 'active'
RETURNING id, user_id, status;

-- name: ReinstateAgent :one
UPDATE agents
SET status = 'active', updated_at = now()
WHERE id = $1 AND status = 'suspended'
RETURNING id, user_id, status;

-- name: SuspendActiveAgentApplications :exec
UPDATE agent_applications
SET status = 'suspended', updated_at = now()
WHERE agent_id = $1 AND status = 'active';

-- name: ReinstateSuspendedAgentApplications :exec
UPDATE agent_applications
SET status = 'active', updated_at = now()
WHERE agent_id = $1 AND status = 'suspended';

-- name: RevokeUserSessions :many
UPDATE sessions
SET revoked_at = now()
WHERE sessions.user_id = $1 AND revoked_at IS NULL
RETURNING id;
