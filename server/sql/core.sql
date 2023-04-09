CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS user_prompts
(
    request_id UUID      NOT NULL PRIMARY KEY,
    user_id    UUID      NOT NULL,
    prompt     TEXT      NOT NULL,
    timestamp  TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS openai_responses
(
    request_id        UUID      NOT NULL PRIMARY KEY REFERENCES user_prompts (request_id),
    user_id           UUID      NOT NULL,
    response_raw      TEXT      NOT NULL,
    response          TEXT      NOT NULL,
    prompt_tokens     SMALLINT  NOT NULL,
    completion_tokens SMALLINT  NOT NULL,
    model_id          TEXT      NOT NULL,
    timestamp         TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS users
(
    user_id         UUID      NOT NULL PRIMARY KEY,
    email           TEXT,
    email_verified  BOOLEAN   NOT NULL DEFAULT FALSE,
    web_fingerprint TEXT,
    is_active       BOOLEAN   NOT NULL DEFAULT FALSE,
    is_premium      BOOLEAN   NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    update_at       TIMESTAMP NOT NULL DEFAULT NOW()
);

INSERT INTO users (user_id, is_active) VALUES ('00000000-0000-0000-0000-000000000000', TRUE);

CREATE TABLE IF NOT EXISTS api_tokens
(
    token      UUID      NOT NULL PRIMARY KEY,
    user_id    UUID      NOT NULL REFERENCES users (user_id),
    is_active  BOOLEAN   NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS ind_user_api_tokens_user_id ON api_tokens (user_id);

CREATE TABLE IF NOT EXISTS successful_requests
(
    request_id UUID PRIMARY KEY NOT NULL REFERENCES user_prompts (request_id),
    user_id    UUID             NOT NULL REFERENCES users (user_id),
    token      UUID,
    timestamp  TIMESTAMP        NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS ind_successful_requests_timestamp ON successful_requests (timestamp);
CREATE INDEX IF NOT EXISTS ind_successful_requests_user_id ON successful_requests (user_id);
