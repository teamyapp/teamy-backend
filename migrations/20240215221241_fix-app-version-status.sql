-- +migrate Up
ALTER table app_version
DROP COLUMN is_ready,
ADD COLUMN status VARCHAR(50) NOT NULL DEFAULT 'INIT';

-- +migrate Down
ALTER table app_version
ADD COLUMN is_ready BOOLEAN NOT NULL DEFAULT FALSE,
DROP COLUMN status;


