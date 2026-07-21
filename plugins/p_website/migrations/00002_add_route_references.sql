-- +goose Up
CREATE TABLE IF NOT EXISTS p_website_route_references (
    db_route_id BIGINT NOT NULL REFERENCES db_routes (id) ON DELETE CASCADE,
    v_node_id   BIGINT NOT NULL REFERENCES filesystem_nodes (id) ON DELETE CASCADE,
    PRIMARY KEY (db_route_id, v_node_id)
);

CREATE INDEX IF NOT EXISTS idx_p_website_route_references_db_route_id ON p_website_route_references (db_route_id);
CREATE INDEX IF NOT EXISTS idx_p_website_route_references_v_node_id ON p_website_route_references (v_node_id);

-- +goose Down
DROP TABLE IF EXISTS p_website_route_references;
