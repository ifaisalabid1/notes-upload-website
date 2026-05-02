CREATE TABLE IF NOT EXISTS notes (
    id          TEXT        PRIMARY KEY,
    subject_id  TEXT        NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    title       TEXT        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    file_key    TEXT        NOT NULL UNIQUE,
    file_type   TEXT        NOT NULL,
    file_size   BIGINT      NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);