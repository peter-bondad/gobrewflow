-- Remove accepted_at column from invitations table
ALTER TABLE invitations
    DROP COLUMN IF EXISTS accepted_at;
