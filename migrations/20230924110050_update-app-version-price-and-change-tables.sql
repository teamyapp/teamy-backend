-- +migrate Up
ALTER TABLE app_version_price
    ADD COLUMN tag VARCHAR(100),
	DROP CONSTRAINT pk_app_version_price;

ALTER TABLE app_version_price
    ADD CONSTRAINT pk_app_version_price PRIMARY KEY (app_id, version_number, currency);


ALTER TABLE app_version_change
    DROP CONSTRAINT pk_app_version_change;

ALTER TABLE app_version_change
    ADD COLUMN change_id BIGINT PRIMARY KEY;

-- +migrate Down
DELETE FROM app_version_change;

ALTER TABLE app_version_change
    DROP CONSTRAINT app_version_change_pkey,
    DROP COLUMN change_id;

ALTER TABLE app_version_change
    ADD CONSTRAINT pk_app_version_change PRIMARY KEY (app_id, version_number);


DELETE FROM app_version_price;
ALTER TABLE app_version_price
    DROP CONSTRAINT pk_app_version_price;

ALTER TABLE app_version_price
    ADD CONSTRAINT pk_app_version_price PRIMARY KEY (app_id, version_number),
    DROP COLUMN tag;
