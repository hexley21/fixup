CREATE TYPE USER_ROLE AS ENUM ('CUSTOMER', 'PROVIDER', 'MODERATOR', 'ADMIN');

CREATE TABLE users (
    id BIGINT PRIMARY KEY NOT NULL,
    first_name VARCHAR(30) NOT NULL,
    last_name VARCHAR(30) NOT NULL,
    phone_number VARCHAR(15) NOT NULL,
    email VARCHAR(40) NOT NULL,
    picture TEXT,
    hash VARCHAR(128) NOT NULL CHECK(LENGTH(hash) = 128),
    role USER_ROLE NOT NULL,
    verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(email)
);

CREATE TABLE providers (
    personal_id_number BYTEA NOT NULL,
    personal_id_preview VARCHAR(5) NOT NULL,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE
);
