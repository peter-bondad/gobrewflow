-- +goose Down
ALTER TABLE invitations
DROP COLUMN cancelled_at;