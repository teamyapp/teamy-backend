-- +migrate Up

UPDATE task
SET status = 'TODO'
WHERE status = 'UPCOMING';

-- +migrate Down
UPDATE task
SET status = 'UPCOMING'
WHERE status = 'TODO';