CREATE TABLE sessions (
    session_key VARCHAR(255) NOT NULL,
    user_id MEDIUMINT,
    PRIMARY KEY (session_key),
    FOREIGN KEY (user_id) REFERENCES users(id)
);
