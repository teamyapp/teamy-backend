-- +migrate Up
ALTER TABLE app
	DROP COLUMN apisecret,
    DROP COLUMN app_name,
	ADD COLUMN api_secret VARCHAR(128) NOT NULL,
	ADD COLUMN name VARCHAR(100) NOT NULL;

-- +migrate Down
ALTER TABLE app
    DROP COLUMN name,
	DROP COLUMN api_secret,
    ADD COLUMN app_name VARCHAR(100) NOT NULL,
    ADD COLUMN apisecret VARCHAR(128) NOT NULL;
