-- +migrate Up
ALTER TABLE task RENAME COLUMN is_planned TO is_scheduled;

-- +migrate Down
ALTER TABLE task RENAME COLUMN is_scheduled TO is_planned;