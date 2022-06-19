CREATE TABLE sessions (
    session_id uuid NOT NULL PRIMARY KEY,
    keyword varchar,
    per int,
    photo_count int,
    created_at TIMESTAMPTZ,
    links varchar[]
);