-- +goose Up
CREATE TABLE inquiries (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    student_user_id uuid NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    agent_offer_id uuid NOT NULL REFERENCES agent_offers (id) ON DELETE RESTRICT,
    submission_id uuid NOT NULL,
    message text NOT NULL,
    status text NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'closed')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT inquiries_message_length CHECK (char_length(btrim(message)) BETWEEN 1 AND 1000),
    CONSTRAINT inquiries_student_submission_unique UNIQUE (student_user_id, submission_id)
);

CREATE INDEX idx_inquiries_agent_offer_status_created_at
    ON inquiries (agent_offer_id, status, created_at DESC);

CREATE INDEX idx_inquiries_student_created_at
    ON inquiries (student_user_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS inquiries;
