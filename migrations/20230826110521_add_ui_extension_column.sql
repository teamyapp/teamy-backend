-- +migrate Up
ALTER TABLE app_version
    ADD has_ui_extension BOOL NOT NULL DEFAULT FALSE;

-- +migrate Down
ALTER TABLE app_version
    DROP has_ui_extension;