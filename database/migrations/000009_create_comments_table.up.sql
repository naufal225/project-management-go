CREATE TABLE IF NOT EXISTS comments (
    internal_id BIGSERIAL PRIMARY KEY,
    public_id UUID NOT NULL DEFAULT gen_random_uuid(),
    card_internal_id BIGINT NOT NULL REFERENCES cards(internal_id) ON DELETE CASCADE,
    user_internal_id BIGINT NOT NULL REFERENCES users(internal_id) ON DELETE CASCADE,
    message TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT commnets_public_id_unique UNIQUE (public_id)
)