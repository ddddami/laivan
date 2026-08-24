-- +goose Up
CREATE TABLE agent_campuses (
    agent_id uuid NOT NULL REFERENCES agents (id) ON DELETE RESTRICT,
    campus_id uuid NOT NULL REFERENCES campuses (id) ON DELETE RESTRICT,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (agent_id, campus_id)
);

CREATE INDEX idx_agent_campuses_campus_id ON agent_campuses (campus_id);

INSERT INTO agent_campuses (agent_id, campus_id)
SELECT DISTINCT ao.agent_id, p.campus_id
FROM agent_offers ao
JOIN property_unit_types put ON put.id = ao.property_unit_type_id
JOIN properties p ON p.id = put.property_id
ON CONFLICT (agent_id, campus_id) DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS agent_campuses;
