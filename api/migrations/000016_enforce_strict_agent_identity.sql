-- +goose Up
-- +goose StatementBegin
DELETE FROM agent_applications WHERE agent_id IN (SELECT id FROM agents WHERE user_id IS NULL);
DELETE FROM agent_offers WHERE agent_id IN (SELECT id FROM agents WHERE user_id IS NULL);
DELETE FROM agent_campuses WHERE agent_id IN (SELECT id FROM agents WHERE user_id IS NULL);
DELETE FROM agents WHERE user_id IS NULL;
ALTER TABLE agents
    DROP CONSTRAINT agents_user_id_fkey,
    ADD CONSTRAINT agents_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE RESTRICT;
ALTER TABLE agents ALTER COLUMN user_id SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE agents ALTER COLUMN user_id DROP NOT NULL;
-- +goose StatementEnd
