CREATE TABLE mutable_drawings (
    id MEDIUMINT NOT NULL AUTO_INCREMENT,
    user_id MEDIUMINT NOT NULL,
    name VARCHAR(100),
    data JSON NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX idx_mut_drawings_name ON mutable_drawings(name);
