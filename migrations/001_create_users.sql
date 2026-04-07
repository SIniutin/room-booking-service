-- +goose Up
CREATE TABLE users (
    id   UUID PRIMARY KEY,
    email VARCHAR(254) NOT NULL UNIQUE,
    role  VARCHAR(10)  NOT NULL CHECK (role IN ('admin', 'user')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE credentials (
    user_id       UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    password_hash TEXT NOT NULL
);

-- +goose Down
DROP TABLE credentials;
DROP TABLE users;
