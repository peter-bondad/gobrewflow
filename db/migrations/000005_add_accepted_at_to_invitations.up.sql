-- Add accepted_at column to invitations table
ALTER TABLE invitations
    ADD COLUMN IF NOT EXISTS accepted_at TIMESTAMPTZ;
