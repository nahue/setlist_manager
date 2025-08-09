-- +goose Up
CREATE TABLE bands (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    created_by TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    is_active INTEGER DEFAULT 1,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE band_members (
    id TEXT PRIMARY KEY,
    band_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'member', -- 'owner', 'admin', 'member'
    joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    is_active INTEGER DEFAULT 1,
    FOREIGN KEY (band_id) REFERENCES bands(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(band_id, user_id)
);

CREATE TABLE band_invitations (
    id TEXT PRIMARY KEY,
    band_id TEXT NOT NULL,
    invited_email TEXT NOT NULL,
    invited_by TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'member',
    status TEXT NOT NULL DEFAULT 'pending', -- 'pending', 'accepted', 'declined', 'expired'
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    accepted_at DATETIME,
    declined_at DATETIME,
    FOREIGN KEY (band_id) REFERENCES bands(id) ON DELETE CASCADE,
    FOREIGN KEY (invited_by) REFERENCES users(id) ON DELETE CASCADE
);

-- Indexes for better performance
CREATE INDEX idx_bands_created_by ON bands(created_by);
CREATE INDEX idx_bands_is_active ON bands(is_active);
CREATE INDEX idx_band_members_band_id ON band_members(band_id);
CREATE INDEX idx_band_members_user_id ON band_members(user_id);
CREATE INDEX idx_band_members_band_user ON band_members(band_id, user_id);
CREATE INDEX idx_band_invitations_band_id ON band_invitations(band_id);
CREATE INDEX idx_band_invitations_invited_email ON band_invitations(invited_email);
CREATE INDEX idx_band_invitations_status ON band_invitations(status);
CREATE INDEX idx_band_invitations_expires_at ON band_invitations(expires_at);

-- +goose Down
DROP INDEX IF EXISTS idx_band_invitations_expires_at;
DROP INDEX IF EXISTS idx_band_invitations_status;
DROP INDEX IF EXISTS idx_band_invitations_invited_email;
DROP INDEX IF EXISTS idx_band_invitations_band_id;
DROP INDEX IF EXISTS idx_band_members_band_user;
DROP INDEX IF EXISTS idx_band_members_user_id;
DROP INDEX IF EXISTS idx_band_members_band_id;
DROP INDEX IF EXISTS idx_bands_is_active;
DROP INDEX IF EXISTS idx_bands_created_by;

DROP TABLE IF EXISTS band_invitations;
DROP TABLE IF EXISTS band_members;
DROP TABLE IF EXISTS bands;
