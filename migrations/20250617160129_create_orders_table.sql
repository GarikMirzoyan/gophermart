-- +goose Up
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    number VARCHAR(255) NOT NULL UNIQUE,
    status VARCHAR(50) NOT NULL,
    accrual INTEGER,
    uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL,
    user_id INTEGER NOT NULL REFERENCES users(id)
);

-- +goose Down
DROP TABLE orders;