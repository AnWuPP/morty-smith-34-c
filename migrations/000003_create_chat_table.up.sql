CREATE TABLE chats (
    id SERIAL PRIMARY KEY,
    chat_id BIGINT NOT NULL UNIQUE,
    campus_name VARCHAR(16) NOT NULL,
    thread_id BIGINT NOT NULL DEFAULT -1,
    created_at TIMESTAMP DEFAULT NOW()
);