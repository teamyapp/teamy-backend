-- +migrate Up
ALTER TABLE app_version 
    ADD COLUMN error_message VARCHAR(100);

-- +migrate Down
ALTER TABLE app_version
    DROP COLUMN error_message;
