-- +migrate Up
ALTER TABLE app_version
    ADD icon_url VARCHAR(2048);

-- +migrate Down
ALTER TABLE app_version
    DROP icon_url;
