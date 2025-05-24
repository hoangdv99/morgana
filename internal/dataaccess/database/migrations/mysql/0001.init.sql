CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED PRIMARY KEY
    username VARCHAR(256)
);

CREATE TABLE IF NOT EXISTS user_passwords (
    user_id BIGINT UNSIGNED PRIMARY KEY,
    hash VARCHAR(128),
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS download_tasks (
    id BIGINT UNSIGNED PRIMARY KEY,
    user_id BIGINT UNSIGNED,
    download_type SMALLINT NOT NULL,
    url TEXT NOT NULL,
    download_status SMALLINT NOT NULL,
    metadata TEXT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
