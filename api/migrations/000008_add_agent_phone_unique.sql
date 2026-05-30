-- +goose Up
-- Prevent duplicate agent identities by enforcing one phone number per agent.
ALTER TABLE agents
ADD CONSTRAINT agents_phone_number_unique UNIQUE (phone_number);

-- +goose Down
ALTER TABLE agents DROP CONSTRAINT IF EXISTS agents_phone_number_unique;
