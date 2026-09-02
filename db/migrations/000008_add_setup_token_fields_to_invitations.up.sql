-- Add setup_token_hash and setup_token_hash_expires_at columns to invitations table
ALTER TABLE invitations
    ADD COLUMN IF NOT EXISTS setup_token_hash TEXT,
    ADD COLUMN IF NOT EXISTS setup_token_hash_expires_at TIMESTAMPTZ;