-- +goose Up
ALTER TABLE media
    ADD COLUMN object_key text,
    ADD COLUMN content_type text,
    ADD COLUMN size_bytes integer CHECK (size_bytes IS NULL OR size_bytes > 0);

ALTER TABLE media
    ADD CONSTRAINT media_exactly_one_target_check CHECK (
        ((property_id IS NOT NULL)::integer +
         (property_unit_type_id IS NOT NULL)::integer +
         (agent_offer_id IS NOT NULL)::integer) = 1
    );

-- +goose Down
ALTER TABLE media
    DROP CONSTRAINT media_exactly_one_target_check,
    DROP COLUMN size_bytes,
    DROP COLUMN content_type,
    DROP COLUMN object_key;
