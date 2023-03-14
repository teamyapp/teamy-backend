-- +migrate Up
ALTER TABLE apps_github_app_installation
    ALTER COLUMN id TYPE INT USING id::INT;

-- +migrate Down
ALTER TABLE apps_github_app_installation
    ALTER COLUMN id TYPE BIGINT USING id::BIGINT;
