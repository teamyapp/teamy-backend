-- +migrate Up
CREATE TABLE apps_github_app_installation
(
	id         VARCHAR(50) NOT NULL,
	team_id    BIGINT NOT NULL REFERENCES team (id) ON UPDATE CASCADE ON DELETE CASCADE,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (id, team_id)
);

CREATE TABLE apps_github_app_install_state
(
	id           BIGINT NOT NULL PRIMARY KEY,
	team_id      BIGINT NOT NULL REFERENCES team (id) ON UPDATE CASCADE ON DELETE CASCADE,
	redirect_url VARCHAR(2048),
	created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- +migrate Down
DROP TABLE apps_github_app_install_state;
DROP TABLE apps_github_app_installation;
