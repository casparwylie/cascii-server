CREATE TABLE immutable_drawings (
    id MEDIUMINT NOT NULL AUTO_INCREMENT,
    short_key VARCHAR(10),
    data JSON,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
);

CREATE INDEX idx_immut_drawings_short_key ON immutable_drawings(short_key);
