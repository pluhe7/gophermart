-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    login VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    balance NUMERIC(8,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TYPE order_status AS ENUM ('NEW','PROCESSING','INVALID','PROCESSED');
CREATE TABLE IF NOT EXISTS orders (
    number VARCHAR(255) PRIMARY KEY,
    user_id SERIAL NOT NULL REFERENCES users (id),
    status order_status DEFAULT 'NEW',
    uploaded_at TIMESTAMP NOT NULL
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TYPE transaction_type AS ENUM ('withdrawal','accrual');
CREATE TABLE IF NOT EXISTS transactions (
     id SERIAL PRIMARY KEY,
     user_id SERIAL NOT NULL REFERENCES users (id),
     order_number VARCHAR(255) NOT NULL,
     sum NUMERIC(8,2) NOT NULL,
     type transaction_type NOT NULL,
     processed_at TIMESTAMP NOT NULL
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS sessions (
    token VARCHAR(64) PRIMARY KEY,
    user_id SERIAL NOT NULL REFERENCES users (id),
    created_at TIMESTAMP NOT NULL,
    expire_at TIMESTAMP NOT NULL,
    last_activity_at TIMESTAMP NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS sessions;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS transactions;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS orders;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
