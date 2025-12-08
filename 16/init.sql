CREATE TABLE IF NOT EXISTS notes (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_notes_title_gin
    ON notes USING GIN (to_tsvector('simple', title));

CREATE INDEX IF NOT EXISTS idx_notes_created_id
    ON notes (created_at, id);

CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_notes_updated ON notes;
CREATE TRIGGER trg_notes_updated
BEFORE UPDATE ON notes
FOR EACH ROW
EXECUTE FUNCTION update_updated_at();
