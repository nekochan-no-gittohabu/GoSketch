CREATE TABLE photos (
    link varchar,
    dur int,
    created_at TIMESTAMPTZ,
    session_id uuid,
    FOREIGN KEY(session_id) REFERENCES sessions(session_id) ON DELETE CASCADE
);