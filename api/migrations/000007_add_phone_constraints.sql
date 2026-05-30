-- +goose Up
-- Enforce E.164 format: +234 followed by exactly 10 digits (14 characters total).
-- This prevents duplicate representations of the same phone number in the database.
ALTER TABLE agents
ADD CONSTRAINT agents_phone_number_format
CHECK (phone_number ~ '^\+234[0-9]{10}$');

ALTER TABLE agents
ADD CONSTRAINT agents_whatsapp_number_format
CHECK (whatsapp_number IS NULL OR whatsapp_number ~ '^\+234[0-9]{10}$');

-- +goose Down
ALTER TABLE agents DROP CONSTRAINT IF EXISTS agents_phone_number_format;
ALTER TABLE agents DROP CONSTRAINT IF EXISTS agents_whatsapp_number_format;
