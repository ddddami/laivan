-- +goose Up
-- Align media.size_bytes with domain model int64 to prevent 32-bit overflow.
ALTER TABLE media ALTER COLUMN size_bytes TYPE bigint;

-- +goose Down
ALTER TABLE media ALTER COLUMN size_bytes TYPE integer;
