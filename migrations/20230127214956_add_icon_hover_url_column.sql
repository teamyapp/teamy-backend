-- +migrate Up
ALTER TABLE task_link
ADD icon_hover_url VARCHAR(2048);

-- +migrate Down
ALTER TABLE task_link
DROP icon_hover_url VARCHAR(2048);
