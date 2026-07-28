-- +goose Up
ALTER TABLE db_routes ADD COLUMN IF NOT EXISTS grapes_project TEXT;

-- +goose Down
ALTER TABLE db_routes DROP COLUMN IF EXISTS grapes_project;
