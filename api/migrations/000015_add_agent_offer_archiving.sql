-- +goose Up
ALTER TABLE agent_offers
    ADD COLUMN archived_at timestamptz;

-- +goose Down
ALTER TABLE agent_offers DROP COLUMN archived_at;
