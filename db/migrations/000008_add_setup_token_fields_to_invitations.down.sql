-- Drop setup_token_hash and setup_token_hash_expires_at columns from invitations table
ALTER TABLE invitations
    DROP COLUMN IF EXISTS setup_token_hash,
    DROP COLUMN IF EXISTS setup_token_hash_expires_at;