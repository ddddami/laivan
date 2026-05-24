-- +goose Up
INSERT INTO campuses (slug, name, short_name)
VALUES ('futa', 'Federal University of Technology, Akure', 'FUTA');

-- +goose Down
DELETE FROM campuses WHERE slug = 'futa';
