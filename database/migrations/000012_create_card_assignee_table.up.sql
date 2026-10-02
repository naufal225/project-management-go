CREATE TABLE IF NOT EXISTS card_assignees (
    card_internal_id BIGINT NOT NULL REFERENCES cards(internal_id) ON DELETE CASCADE,
    user_internal_id BIGINT NOT NULL REFERENCES users(internal_id) ON DELETE CASCADE,
    PRIMARY KEY (card_internal_id, user_internal_id)
);