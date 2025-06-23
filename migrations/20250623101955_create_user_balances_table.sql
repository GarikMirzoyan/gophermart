-- +goose Up
CREATE TABLE user_balances (
	user_id INTEGER PRIMARY KEY,
	current_balance NUMERIC NOT NULL DEFAULT 0,
	total_withdrawn NUMERIC NOT NULL DEFAULT 0
);

-- +goose Down
DROP TABLE user_balances;
