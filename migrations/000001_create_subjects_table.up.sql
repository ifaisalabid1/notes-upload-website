CREATE TABLE IF NOT EXISTS subjects (
    id          TEXT      PRIMARY KEY,
    name        TEXT      NOT NULL,
    slug        TEXT      NOT NULL UNIQUE,
    description TEXT      NOT NULL DEFAULT '',
    created_at  DATETIME  NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at  DATETIME  NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);