-- +goose Up
CREATE TABLE IF NOT EXISTS withdrawals (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    order_number TEXT NOT NULL UNIQUE,
    sum NUMERIC(18, 2) NOT NULL CHECK (sum >= 0),
    processed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_withdrawals_user_id_processed_at ON withdrawals(user_id, processed_at DESC);

-- +goose Down
DROP TABLE withdrawals;
