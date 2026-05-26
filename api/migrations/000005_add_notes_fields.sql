-- +goose Up
ALTER TABLE property_unit_types ADD COLUMN notes text;
ALTER TABLE agent_offers ADD COLUMN notes text;

-- +goose Down
ALTER TABLE property_unit_types DROP COLUMN notes;
ALTER TABLE agent_offers DROP COLUMN notes;
