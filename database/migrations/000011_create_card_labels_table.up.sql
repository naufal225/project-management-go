CREATE TABLE IF NOT EXISTS card_labels (
    card_internal_id BIGINT NOT NULL REFERENCES cards(internal_id) ON DELETE CASCADE,
    label_internal_id BIGINT NOT NULL REFERENCES labels(internal_id) ON DELETE CASCADE,
    PRIMARY KEY (card_internal_id, label_internal_id)
);