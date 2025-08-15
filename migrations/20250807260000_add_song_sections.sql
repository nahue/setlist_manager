-- +goose Up
CREATE TABLE song_sections (
    id TEXT PRIMARY KEY,
    song_id TEXT NOT NULL,
    title TEXT NOT NULL,
    key TEXT,
    body TEXT,
    position INTEGER NOT NULL DEFAULT 0,
    created_by TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN NOT NULL DEFAULT 1,
    FOREIGN KEY (song_id) REFERENCES songs(id) ON DELETE CASCADE,
    FOREIGN KEY (created_by) REFERENCES users(id)
);

-- Create indexes for better performance
CREATE INDEX idx_song_sections_song_id ON song_sections(song_id, is_active);
CREATE INDEX idx_song_sections_position ON song_sections(song_id, position);

-- +goose Down
DROP INDEX IF EXISTS idx_song_sections_position;
DROP INDEX IF EXISTS idx_song_sections_song_id;
DROP TABLE IF EXISTS song_sections;
