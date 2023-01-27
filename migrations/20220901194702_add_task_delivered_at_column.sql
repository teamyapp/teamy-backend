-- +migrate Up
ALTER TABLE task
ADD delivered_at TIMESTAMP;

UPDATE task
SET delivered_at = updated_at
WHERE status = 'DELIVERED';

-- +migrate Down
ALTER TABLE task
DROP delivered_at TIMESTAMP;