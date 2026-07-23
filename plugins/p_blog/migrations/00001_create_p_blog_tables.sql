-- +goose Up
CREATE EXTENSION IF NOT EXISTS ltree;

CREATE TABLE IF NOT EXISTS blog_tags (
    id         BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    name       ltree NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_blog_tags_deleted_at ON blog_tags (deleted_at);
CREATE INDEX IF NOT EXISTS idx_blog_tags_name ON blog_tags USING gist (name);

CREATE TABLE IF NOT EXISTS blogs (
    id            BIGSERIAL PRIMARY KEY,
    created_at    TIMESTAMPTZ,
    updated_at    TIMESTAMPTZ,
    deleted_at    TIMESTAMPTZ,
    title         TEXT NOT NULL,
    slug          VARCHAR(255) NOT NULL,
    description   TEXT,
    created_by_id BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    content       TEXT
);

CREATE INDEX IF NOT EXISTS idx_blogs_deleted_at ON blogs (deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_blogs_slug ON blogs (slug) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_blogs_created_by_id ON blogs (created_by_id);

CREATE TABLE IF NOT EXISTS p_blog_tags (
    blog_id     BIGINT NOT NULL REFERENCES blogs (id) ON DELETE CASCADE,
    blog_tag_id BIGINT NOT NULL REFERENCES blog_tags (id) ON DELETE CASCADE,
    PRIMARY KEY (blog_id, blog_tag_id)
);

-- +goose Down
DROP TABLE IF EXISTS p_blog_tags;
DROP TABLE IF EXISTS blogs;
DROP TABLE IF EXISTS blog_tags;
