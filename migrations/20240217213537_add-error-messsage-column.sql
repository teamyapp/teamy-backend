-- +migrate Up
ALTER TABLE app_version 
    ADD COLUMN error_message VARCHAR(100),
    DROP COLUMN icon_url;

-- +migrate Down
ALTER TABLE app_version
    ADD COLUMN icon_url VARCHAR(2048),
    DROP COLUMN error_message;
