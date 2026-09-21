-- +goose Up
ALTER TABLE media
    ADD COLUMN removed_at timestamptz,
    ADD COLUMN removed_by_user_id uuid REFERENCES users (id) ON DELETE RESTRICT;

-- +goose Down
ALTER TABLE media
    DROP COLUMN removed_by_user_id,
    DROP COLUMN removed_at;
