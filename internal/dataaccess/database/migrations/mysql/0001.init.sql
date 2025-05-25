CREATE TABLE IF NOT EXISTS accounts (
    id BIGINT UNSIGNED PRIMARY KEY
    account_name VARCHAR(256)
);

CREATE TABLE IF NOT EXISTS account_passwords (
    account_id BIGINT UNSIGNED PRIMARY KEY,
    hash VARCHAR(128),
    FOREIGN KEY (account_id) REFERENCES accounts(id)
);

CREATE TABLE IF NOT EXISTS token_public_keys (
    id BIGINT UNSIGNED PRIMARY KEY,
    public_key VARBINARY(4096) NOT NULL
);

CREATE TABLE IF NOT EXISTS download_tasks (
    id BIGINT UNSIGNED PRIMARY KEY,
    account_id BIGINT UNSIGNED,
    download_type SMALLINT NOT NULL,
    url TEXT NOT NULL,
    download_status SMALLINT NOT NULL,
    metadata TEXT NOT NULL,
    FOREIGN KEY (account_id) REFERENCES accounts(id)
);
