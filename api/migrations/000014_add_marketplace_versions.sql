-- +goose Up
ALTER TABLE properties
    ADD COLUMN version integer NOT NULL DEFAULT 1 CHECK (version > 0);

ALTER TABLE property_unit_types
    ADD COLUMN version integer NOT NULL DEFAULT 1 CHECK (version > 0);

ALTER TABLE agent_offers
    ADD COLUMN version integer NOT NULL DEFAULT 1 CHECK (version > 0);

-- +goose Down
ALTER TABLE agent_offers DROP COLUMN version;
ALTER TABLE property_unit_types DROP COLUMN version;
ALTER TABLE properties DROP COLUMN version;
