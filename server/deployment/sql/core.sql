CREATE TABLE IF NOT EXISTS user_prompt (
    id         SERIAL8,
    request_id TEXT NOT NULL,
    user_id    TEXT NOT NULL DEFAULT 'NA',
    prompt     TEXT NOT NULL,
    timestamp  TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS openai_response (
    id         SERIAL8,
    request_id TEXT NOT NULL,
    user_id    TEXT NOT NULL DEFAULT 'NA',
    response   TEXT NOT NULL,
    timestamp  TIMESTAMP NOT NULL DEFAULT NOW()
);
