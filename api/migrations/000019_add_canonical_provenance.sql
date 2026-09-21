-- +goose Up
ALTER TABLE properties
    ADD COLUMN created_by_user_id uuid REFERENCES users (id) ON DELETE RESTRICT;

ALTER TABLE property_unit_types
    ADD COLUMN created_by_user_id uuid REFERENCES users (id) ON DELETE RESTRICT;

-- +goose Down
ALTER TABLE property_unit_types DROP COLUMN created_by_user_id;
ALTER TABLE properties DROP COLUMN created_by_user_id;
