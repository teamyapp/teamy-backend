-- +migrate Up
ALTER TABLE team ADD COLUMN max_group_order_index INT NOT NULL DEFAULT 0;

-- +migrate Down
ALTER TABLE team DROP COLUMN max_group_order_index;
