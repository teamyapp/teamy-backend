-- +migrate Up
ALTER TABLE task_link ADD COLUMN preview_title VARCHAR(255) NOT NULL DEFAULT '';

-- +migrate Down
ALTER TABLE task_link DROP COLUMN preview_title;
