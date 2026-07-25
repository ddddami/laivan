-- +goose Up
CREATE INDEX idx_properties_campus_id ON properties (campus_id);

-- +goose Down
DROP INDEX idx_properties_campus_id;
