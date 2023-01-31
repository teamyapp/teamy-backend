-- +migrate Up
ALTER TABLE app
	ADD description TEXT,
    ADD app_name VARCHAR(100) NOT NULL DEFAULT '';

ALTER TABLE app_version
    ADD changes TEXT,
    DROP COLUMN app_name;

DROP TABLE app_installation;

CREATE TABLE app_team_installation (
	app_id BIGINT NOT NULL REFERENCES app (id) ON UPDATE CASCADE ON DELETE CASCADE,
	installed_team_id BIGINT NOT NULL REFERENCES team (id) ON UPDATE CASCADE ON DELETE CASCADE,
	installed_by_user_id BIGINT NOT NULL REFERENCES "user" (id) ON UPDATE CASCADE ON DELETE SET NULL,
	enabled_version_number INT NOT NULL,
	installed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (app_id, installed_team_id)
);

-- +migrate Down
DROP TABLE app_team_installation;

CREATE TABLE app_installation (
	id BIGINT NOT NULL PRIMARY KEY,
	app_id BIGINT NOT NULL REFERENCES app (id) ON UPDATE CASCADE ON DELETE CASCADE,
	installed_team_id BIGINT NOT NULL REFERENCES team (id) ON UPDATE CASCADE ON DELETE CASCADE,
	installed_by_user_id BIGINT NOT NULL REFERENCES "user" (id) ON UPDATE CASCADE ON DELETE CASCADE,
	installed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE app_version
	ADD app_name VARCHAR(100) NOT NULL default '',
	DROP COLUMN app_name;

ALTER TABLE app
	DROP COLUMN description,
	DROP COLUMN app_name;
