-- +goose Up
ALTER TABLE blogs ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE blogs ADD COLUMN IF NOT EXISTS slug VARCHAR(255);

UPDATE blogs SET slug = 'blog-' || id WHERE slug IS NULL OR slug = '';

ALTER TABLE blogs ALTER COLUMN slug SET NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_blogs_slug ON blogs (slug) WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_blogs_slug;
ALTER TABLE blogs DROP COLUMN IF EXISTS slug;
ALTER TABLE blogs DROP COLUMN IF EXISTS description;
