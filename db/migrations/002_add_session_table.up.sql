CREATE TABLE sessions (
    `key` VARCHAR(255) NOT NULL,
    user_id MEDIUMINT,
    PRIMARY KEY (`key`),
    FOREIGN KEY (user_id) REFERENCES users(id)
);
