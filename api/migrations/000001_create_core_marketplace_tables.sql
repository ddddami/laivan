-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE properties (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    area text NOT NULL,
    landmark text,
    description text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE room_types (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id uuid NOT NULL REFERENCES properties (id) ON DELETE CASCADE,
    name text NOT NULL,
    description text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE agents (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    display_name text NOT NULL,
    phone_number text NOT NULL,
    whatsapp_number text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE agent_offers (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    room_type_id uuid NOT NULL REFERENCES room_types (id) ON DELETE CASCADE,
    agent_id uuid NOT NULL REFERENCES agents (id) ON DELETE CASCADE,
    title text NOT NULL,
    description text,
    price_kobo integer NOT NULL CHECK (price_kobo > 0),
    status text NOT NULL DEFAULT 'available' CHECK (status IN ('available', 'unavailable', 'paused')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (room_type_id, agent_id)
);

CREATE TABLE media (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id uuid REFERENCES properties (id) ON DELETE CASCADE,
    room_type_id uuid REFERENCES room_types (id) ON DELETE CASCADE,
    agent_offer_id uuid REFERENCES agent_offers (id) ON DELETE CASCADE,
    uploaded_by_agent_id uuid REFERENCES agents (id) ON DELETE SET NULL,
    url text NOT NULL,
    kind text NOT NULL DEFAULT 'image' CHECK (kind IN ('image', 'video')),
    caption text,
    created_at timestamptz NOT NULL DEFAULT now(),
    CHECK (property_id IS NOT NULL OR room_type_id IS NOT NULL OR agent_offer_id IS NOT NULL)
);

CREATE INDEX idx_room_types_property_id ON room_types (property_id);
CREATE INDEX idx_agent_offers_room_type_id ON agent_offers (room_type_id);
CREATE INDEX idx_agent_offers_agent_id ON agent_offers (agent_id);
CREATE INDEX idx_agent_offers_status ON agent_offers (status);
CREATE INDEX idx_media_property_id ON media (property_id);
CREATE INDEX idx_media_room_type_id ON media (room_type_id);
CREATE INDEX idx_media_agent_offer_id ON media (agent_offer_id);

-- +goose Down
DROP TABLE IF EXISTS media;
DROP TABLE IF EXISTS agent_offers;
DROP TABLE IF EXISTS agents;
DROP TABLE IF EXISTS room_types;
DROP TABLE IF EXISTS properties;
DROP EXTENSION IF EXISTS pgcrypto;
