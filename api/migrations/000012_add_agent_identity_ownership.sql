-- +goose Up
ALTER TABLE agents
    ADD COLUMN user_id uuid REFERENCES users (id) ON DELETE SET NULL,
    ADD COLUMN status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended'));

CREATE UNIQUE INDEX agents_user_id_unique
    ON agents (user_id)
    WHERE user_id IS NOT NULL;

ALTER TABLE agent_offers
    DROP CONSTRAINT agent_offers_agent_id_fkey,
    ADD CONSTRAINT agent_offers_agent_id_fkey
        FOREIGN KEY (agent_id) REFERENCES agents (id) ON DELETE RESTRICT;

CREATE TABLE global_admin_roles (
    user_id uuid PRIMARY KEY REFERENCES users (id) ON DELETE RESTRICT,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE campus_operators (
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    campus_id uuid NOT NULL REFERENCES campuses (id) ON DELETE RESTRICT,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, campus_id)
);

CREATE INDEX idx_campus_operators_campus_id ON campus_operators (campus_id);

CREATE TABLE agent_applications (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    applicant_user_id uuid NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    campus_id uuid NOT NULL REFERENCES campuses (id) ON DELETE RESTRICT,
    name text NOT NULL,
    phone_number text NOT NULL CHECK (phone_number ~ '^\+234[0-9]{10}$'),
    status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'active', 'declined', 'suspended')),
    reviewer_user_id uuid REFERENCES users (id) ON DELETE SET NULL,
    agent_id uuid REFERENCES agents (id) ON DELETE RESTRICT,
    operator_note text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    decided_at timestamptz
);

CREATE UNIQUE INDEX agent_applications_pending_user_campus_unique
    ON agent_applications (applicant_user_id, campus_id)
    WHERE status = 'pending';

CREATE INDEX idx_agent_applications_applicant_user_id ON agent_applications (applicant_user_id);
CREATE INDEX idx_agent_applications_campus_status ON agent_applications (campus_id, status, created_at DESC);

CREATE TABLE audit_events (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_user_id uuid NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    action text NOT NULL,
    resource_type text NOT NULL,
    resource_id uuid NOT NULL,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_events_resource ON audit_events (resource_type, resource_id, created_at DESC);
CREATE INDEX idx_audit_events_actor ON audit_events (actor_user_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS audit_events;
DROP TABLE IF EXISTS agent_applications;
DROP TABLE IF EXISTS campus_operators;
DROP TABLE IF EXISTS global_admin_roles;

ALTER TABLE agent_offers
    DROP CONSTRAINT agent_offers_agent_id_fkey,
    ADD CONSTRAINT agent_offers_agent_id_fkey
        FOREIGN KEY (agent_id) REFERENCES agents (id) ON DELETE CASCADE;

DROP INDEX IF EXISTS agents_user_id_unique;
ALTER TABLE agents DROP COLUMN IF EXISTS status, DROP COLUMN IF EXISTS user_id;
