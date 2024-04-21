-- +migrate Up
ALTER TABLE task ADD COLUMN is_planned BOOLEAN NOT NULL DEFAULT FALSE;

-- +migrate Down
ALTER TABLE task DROP COLUMN is_planned;
