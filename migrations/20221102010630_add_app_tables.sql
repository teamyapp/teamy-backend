-- +migrate Up
CREATE TABLE app (
    id BIGINT NOT NULL PRIMARY KEY,
    apiSecret VARCHAR(128) NOT NULL,
    active_version_number INT,
	installation_count BIGINT NOT NULL DEFAULT 0,
	creator_user_id BIGINT NOT NULL REFERENCES "user" (id) ON UPDATE CASCADE ON DELETE CASCADE,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP
);

CREATE TABLE app_version (
    app_id BIGINT NOT NULL REFERENCES app (id) ON UPDATE CASCADE ON DELETE CASCADE,
	version_number INT NOT NULL,
	app_name VARCHAR(100) NOT NULL,
	icon_url VARCHAR(2048),
	has_ui_extension BOOL NOT NULL DEFAULT FALSE,
	ui_extension_entrypoint_path VARCHAR(4096),
	is_public BOOL NOT NULL DEFAULT FALSE,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP,
	PRIMARY KEY (app_id, version_number)
);

CREATE TABLE app_version_visible_team (
    app_id BIGINT NOT NULL REFERENCES app (id) ON UPDATE CASCADE ON DELETE CASCADE,
    version_number INT NOT NULL,
    team_id BIGINT NOT NULL REFERENCES team (id) ON UPDATE CASCADE ON DELETE CASCADE,
	PRIMARY KEY (app_id, version_number, team_id)
);

CREATE TABLE app_installation (
    id BIGINT NOT NULL PRIMARY KEY,
    app_id BIGINT NOT NULL REFERENCES app (id) ON UPDATE CASCADE ON DELETE CASCADE,
    installed_team_id BIGINT NOT NULL REFERENCES team (id) ON UPDATE CASCADE ON DELETE CASCADE,
	installed_by_user_id BIGINT NOT NULL REFERENCES "user" (id) ON UPDATE CASCADE ON DELETE CASCADE,
	installed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +migrate Down
DROP TABLE app_installation;
DROP TABLE app_version_visible_team;
DROP TABLE app_version;
DROP TABLE app;
