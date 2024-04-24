CREATE DATABASE test_db;

CREATE USER test_user WITH PASSWORD "test_password"

GRANT ALL PRIVILEGES ON DATABASE test_db TO test_user;

CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    message TEXT NOT NULL,
    user_id INT NOT NULL,
    create_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    solved BOOLEAN DEFAULT FALSE
);

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    is_engineer BOOLEAN DEFAULT FALSE
);

CREATE TABLE sessions (
    session_id VARCHAR(128) PRIMARY KEY,
    user_id INT NOT NULL,
    exp_at TIMESTAMP NOT NULL,
    is_engineer BOOLEAN DEFAULT FALSE
);

INSERT INTO users (email, password, is_engineer) 
VALUES ('engineer@engineer.com', 'engineer', true),
('specialist@engineer.com', 'specialist', false);