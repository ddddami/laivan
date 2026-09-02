-- +goose Up
ALTER TABLE agents
    DROP CONSTRAINT agents_user_id_fkey,
    ADD CONSTRAINT agents_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE RESTRICT;

-- +goose Down
ALTER TABLE agents
    DROP CONSTRAINT agents_user_id_fkey,
    ADD CONSTRAINT agents_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE SET NULL;
