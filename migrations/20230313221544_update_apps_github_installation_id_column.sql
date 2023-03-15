-- +migrate Up
ALTER TABLE apps_github_app_installation
    ALTER COLUMN id TYPE INT USING id::INT;

-- +migrate Down
ALTER TABLE apps_github_app_installation
    ALTER COLUMN id TYPE VARCHAR(50) USING id::VARCHAR(50);
