-- +migrate Up
ALTER TABLE apps_github_app_installation
    ALTER COLUMN id TYPE BIGINT USING id::BIGINT;

-- +migrate Down
ALTER TABLE apps_github_app_installation
    ALTER COLUMN id TYPE VARCHAR(50) USING id::VARCHAR(50);
