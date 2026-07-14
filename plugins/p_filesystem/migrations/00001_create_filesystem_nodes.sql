-- +goose Up
CREATE TABLE IF NOT EXISTS filesystem_nodes (
    id           BIGSERIAL PRIMARY KEY,
    created_at   TIMESTAMPTZ,
    updated_at   TIMESTAMPTZ,
    deleted_at   TIMESTAMPTZ,
    name         TEXT NOT NULL,
    is_directory BOOLEAN NOT NULL,
    file_path    TEXT,
    parent_id    BIGINT REFERENCES filesystem_nodes (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_filesystem_nodes_deleted_at ON filesystem_nodes (deleted_at);
CREATE INDEX IF NOT EXISTS idx_filesystem_nodes_parent_id ON filesystem_nodes (parent_id);

DROP INDEX IF EXISTS filesystem_nodes_parent_name_dir_uidx;
CREATE UNIQUE INDEX filesystem_nodes_parent_name_dir_uidx ON filesystem_nodes (
    COALESCE(parent_id, 0),
    name,
    is_directory
) WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS filesystem_nodes;
