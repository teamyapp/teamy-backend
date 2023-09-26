-- +migrate Up
ALTER TABLE app_version_price
	DROP CONSTRAINT pk_app_version_price;

ALTER TABLE app_version_price
    ADD CONSTRAINT pk_app_version_price PRIMARY KEY (app_id, version_number, currency, amount);

ALTER TABLE app_version_change
    DROP CONSTRAINT pk_app_version_change;

ALTER TABLE app_version_change
    ADD CONSTRAINT pk_app_version_change PRIMARY KEY (app_id, version_number, change);

-- +migrate Down
ALTER TABLE app_version_change
    DROP CONSTRAINT pk_app_version_change;

ALTER TABLE app_version_change
    ADD CONSTRAINT pk_app_version_change PRIMARY KEY (app_id, version_number);

ALTER TABLE app_version_price
    DROP CONSTRAINT pk_app_version_price;

ALTER TABLE app_version_price
    ADD CONSTRAINT pk_app_version_price PRIMARY KEY (app_id, version_number);
