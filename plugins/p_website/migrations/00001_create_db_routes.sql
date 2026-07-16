-- +goose Up
CREATE EXTENSION IF NOT EXISTS ltree;

CREATE TABLE IF NOT EXISTS db_routes (
    id          BIGSERIAL PRIMARY KEY,
    created_at  TIMESTAMPTZ,
    updated_at  TIMESTAMPTZ,
    deleted_at  TIMESTAMPTZ,
    path        TEXT NOT NULL UNIQUE,
    ltree_path  ltree GENERATED ALWAYS AS (
        CASE 
            WHEN trim(both '/' from path) = '' THEN 'root'::ltree 
            ELSE replace(replace(replace(trim(both '/' from path), '-', '_'), '.', '_'), '/', '.')::ltree 
        END
    ) STORED,
    page_id     BIGINT NOT NULL REFERENCES filesystem_nodes (id) ON DELETE RESTRICT,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    model       TEXT
);

CREATE INDEX IF NOT EXISTS idx_db_routes_deleted_at ON db_routes (deleted_at);
CREATE INDEX IF NOT EXISTS idx_db_routes_ltree_path ON db_routes USING gist (ltree_path);

-- +goose Down
DROP TABLE IF EXISTS db_routes;
