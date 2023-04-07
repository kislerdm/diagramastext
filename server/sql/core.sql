CREATE TABLE IF NOT EXISTS user_prompt (
    id         SERIAL8,
    request_id TEXT NOT NULL,
    user_id    TEXT NOT NULL DEFAULT 'NA',
    prompt     TEXT NOT NULL,
    timestamp  TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS openai_response (
    id                SERIAL8,
    request_id        TEXT NOT NULL,
    user_id           TEXT NOT NULL DEFAULT 'NA',
    response          TEXT NOT NULL,
    prompt_tokens     SMALLINT NOT NULL,
    completion_tokens SMALLINT NOT NULL,
    model             TEXT NOT NULL,
    timestamp         TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS "user" (
    id              UUID NOT NULL,
    email           TEXT,
    email_verified  BOOLEAN NOT NULL DEFAULT FALSE,
    web_fingerprint TEXT,
    is_active       BOOLEAN NOT NULL DEFAULT FALSE,
    is_premium      BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    update_at       TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_api_token (
   user_id      UUID NOT NULL,
   token        UUID NOT NULL,
   is_active    BOOLEAN NOT NULL DEFAULT TRUE,
   last_used_at TIMESTAMP NOT NULL DEFAULT NOW(),
   created_at   TIMESTAMP NOT NULL DEFAULT NOW()
);
