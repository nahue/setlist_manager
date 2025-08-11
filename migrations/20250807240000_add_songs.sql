-- +goose Up
CREATE TABLE songs (
    id TEXT PRIMARY KEY,
    band_id TEXT NOT NULL,
    title TEXT NOT NULL,
    artist TEXT,
    key TEXT,
    tempo INTEGER,
    notes TEXT,
    created_by TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    is_active INTEGER DEFAULT 1,
    FOREIGN KEY (band_id) REFERENCES bands(id) ON DELETE CASCADE,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE
);

-- Indexes for better performance
CREATE INDEX idx_songs_band_id ON songs(band_id);
CREATE INDEX idx_songs_created_by ON songs(created_by);
CREATE INDEX idx_songs_is_active ON songs(is_active);
CREATE INDEX idx_songs_title ON songs(title);
CREATE INDEX idx_songs_artist ON songs(artist);

-- +goose Down
DROP INDEX IF EXISTS idx_songs_artist;
DROP INDEX IF EXISTS idx_songs_title;
DROP INDEX IF EXISTS idx_songs_is_active;
DROP INDEX IF EXISTS idx_songs_created_by;
DROP INDEX IF EXISTS idx_songs_band_id;

DROP TABLE IF EXISTS songs;
