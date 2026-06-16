-- +goose Up
CREATE TABLE users (
    uuid UUID PRIMARY KEY,
    login VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP,

    CONSTRAINT uq_users_login UNIQUE (login)
);

-- +goose Down
DROP TABLE IF EXISTS users;