-- +goose Up
CREATE TABLE IF NOT EXISTS roles (
    id         BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    name       TEXT UNIQUE
);

CREATE INDEX IF NOT EXISTS idx_roles_deleted_at ON roles (deleted_at);

CREATE TABLE IF NOT EXISTS users (
    id            BIGSERIAL PRIMARY KEY,
    created_at    TIMESTAMPTZ,
    updated_at    TIMESTAMPTZ,
    deleted_at    TIMESTAMPTZ,
    name          TEXT NOT NULL,
    email         TEXT UNIQUE,
    phone         TEXT UNIQUE,
    is_superuser  BOOLEAN NOT NULL DEFAULT false,
    role_id       BIGINT NOT NULL REFERENCES roles (id),
    password      BYTEA,
    password_salt BYTEA,
    timezone      TEXT DEFAULT 'Asia/Kolkata'
);

CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users (deleted_at);
CREATE INDEX IF NOT EXISTS idx_users_role_id ON users (role_id);

-- +goose Down
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS roles;
