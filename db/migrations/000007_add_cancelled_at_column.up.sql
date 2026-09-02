-- +goose Up
ALTER TABLE invitations
ADD COLUMN cancelled_at TIMESTAMP NULL;