-- +goose Up
ALTER TABLE songs ADD COLUMN position INTEGER DEFAULT 0;

-- Update existing songs to have sequential positions based on creation time
UPDATE songs SET position = (
    SELECT COUNT(*) FROM songs s2 
    WHERE s2.band_id = songs.band_id 
    AND s2.created_at <= songs.created_at
);

-- Create index for better performance when ordering by position
CREATE INDEX idx_songs_position ON songs(band_id, position);

-- +goose Down
DROP INDEX IF EXISTS idx_songs_position;
ALTER TABLE songs DROP COLUMN position;
