-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION path_to_ltree(p text) RETURNS ltree AS $$
DECLARE
    cleaned text;
BEGIN
    cleaned := trim(both '/' from p);
    IF cleaned = '' THEN
        RETURN 'root'::ltree;
    END IF;
    cleaned := replace(replace(cleaned, '/', '.'), '-', '_');
    cleaned := regexp_replace(cleaned, '[^a-zA-Z0-9_\.]', '', 'g');
    cleaned := regexp_replace(cleaned, '\.\.', '.', 'g');
    cleaned := trim(both '.' from cleaned);
    IF cleaned = '' THEN
        RETURN 'root'::ltree;
    END IF;
    RETURN cleaned::ltree;
EXCEPTION WHEN OTHERS THEN
    RETURN 'root'::ltree;
END;
$$ LANGUAGE plpgsql IMMUTABLE;
-- +goose StatementEnd

ALTER TABLE db_routes DROP COLUMN IF EXISTS ltree_path;
ALTER TABLE db_routes ADD COLUMN ltree_path ltree GENERATED ALWAYS AS (path_to_ltree(path)) STORED;
CREATE INDEX IF NOT EXISTS idx_db_routes_ltree_path ON db_routes USING gist (ltree_path);

-- +goose Down
ALTER TABLE db_routes DROP COLUMN IF EXISTS ltree_path;
DROP FUNCTION IF EXISTS path_to_ltree(text);
