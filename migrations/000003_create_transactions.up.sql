CREATE TABLE transactions (
    id UUID PRIMARY KEY,
    sender_id UUID NOT NULL REFERENCES users(id),
    receiver_id UUID NOT NULL REFERENCES users(id),
    amount BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);