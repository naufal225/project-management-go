CREATE TABLE IF NOT EXISTS card_attachments (
    internal_id BIGSERIAL PRIMARY KEY,
    public_id UUID NOT NULL DEFAULT gen_random_uuid(),
    file VARCHAR(255) NOT NULL,
    user_internal_id BIGINT NOT NULL REFERENCES users(internal_id) ON DELETE CASCADE,
    card_internal_id BIGINT NOT NULL REFERENCES cards(internal_id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT card_attachment_public_id_unique UNIQUE (public_id)
);