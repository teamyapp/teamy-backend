-- +migrate Up
ALTER TABLE rollout
    ADD "name" VARCHAR(50);

ALTER TABLE app_secret
    ADD "secret" VARCHAR(150);

ALTER TABLE version_selector
    ADD created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	ADD updated_at TIMESTAMP;

-- +migrate Down
ALTER TABLE rollout
    DROP "name";

ALTER TABLE app_secret
    DROP "secret";

ALTER TABLE version_selector
    DROP created_at,
    DROP updated_at;
