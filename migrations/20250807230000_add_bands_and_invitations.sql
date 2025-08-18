-- +goose Up
CREATE TABLE bands (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    created_by TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN NOT NULL DEFAULT 1,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE band_members (
    id TEXT PRIMARY KEY,
    band_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'member',
    joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN NOT NULL DEFAULT 1,
    FOREIGN KEY (band_id) REFERENCES bands(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(band_id, user_id)
);

CREATE TABLE invitations (
    id TEXT PRIMARY KEY,
    band_id TEXT NOT NULL,
    email TEXT NOT NULL,
    token TEXT UNIQUE NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    accepted_at TIMESTAMP,
    declined_at TIMESTAMP,
    FOREIGN KEY (band_id) REFERENCES bands(id) ON DELETE CASCADE
);

-- Create indexes for better performance
CREATE INDEX idx_bands_created_by ON bands(created_by);
CREATE INDEX idx_bands_is_active ON bands(is_active);
CREATE INDEX idx_band_members_band_id ON band_members(band_id, is_active);
CREATE INDEX idx_band_members_user_id ON band_members(user_id, is_active);
CREATE INDEX idx_invitations_band_id ON invitations(band_id);
CREATE INDEX idx_invitations_email ON invitations(email);
CREATE INDEX idx_invitations_token ON invitations(token);
CREATE INDEX idx_invitations_status ON invitations(status);

-- +goose Down
DROP INDEX IF EXISTS idx_invitations_status;
DROP INDEX IF EXISTS idx_invitations_token;
DROP INDEX IF EXISTS idx_invitations_email;
DROP INDEX IF EXISTS idx_invitations_band_id;
DROP INDEX IF EXISTS idx_band_members_user_id;
DROP INDEX IF EXISTS idx_band_members_band_id;
DROP INDEX IF EXISTS idx_bands_is_active;
DROP INDEX IF EXISTS idx_bands_created_by;
DROP TABLE IF EXISTS invitations;
DROP TABLE IF EXISTS band_members;
DROP TABLE IF EXISTS bands;

