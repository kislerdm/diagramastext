CREATE EXTENSION "uuid-ossp";

CREATE TABLE IF NOT EXISTS user_prompt (
    id SERIAL8,
    request_id TEXT,
    user_id TEXT,
    prompt TEXT,
    timestamp TIMESTAMP
);

CREATE TABLE IF NOT EXISTS openai_response (
   id SERIAL8,
   request_id TEXT,
   user_id TEXT,
   response TEXT,
   timestamp TIMESTAMP
);
