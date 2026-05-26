-- +goose Up
ALTER TABLE property_unit_types
  ADD COLUMN category text NOT NULL DEFAULT 'other',
  ADD COLUMN bedroom_count integer CHECK (bedroom_count IS NULL OR bedroom_count >= 0),
  ADD COLUMN has_parlour boolean,
  ADD COLUMN bathroom_type text CHECK (bathroom_type IS NULL OR bathroom_type IN ('private', 'shared', 'unknown')),
  ADD COLUMN kitchen_type text CHECK (kitchen_type IS NULL OR kitchen_type IN ('private', 'shared', 'none', 'unknown')),
  ADD CONSTRAINT property_unit_types_category_check CHECK (category IN ('single_room', 'self_contained', 'room_and_parlour', 'one_bedroom_flat', 'two_bedroom_flat', 'three_bedroom_flat', 'other'));

-- +goose Down
ALTER TABLE property_unit_types
  DROP CONSTRAINT property_unit_types_category_check,
  DROP COLUMN kitchen_type,
  DROP COLUMN bathroom_type,
  DROP COLUMN has_parlour,
  DROP COLUMN bedroom_count,
  DROP COLUMN category;
