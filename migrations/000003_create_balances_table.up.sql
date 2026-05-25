CREATE TABLE IF NOT EXISTS balances (
    id SERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    amount FLOAT NOT NULL
);

CREATE INDEX idx_balances_user_id ON balances(user_id);