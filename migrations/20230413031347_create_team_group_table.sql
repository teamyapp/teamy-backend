-- +migrate Up
CREATE TABLE team_group (
	team_id BIGINT NOT NULL REFERENCES team (id),
	label VARCHAR(20) NOT NULL,
	user_group_id BIGINT NOT NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	CONSTRAINT pk_team_roles PRIMARY KEY (team_id, label)
);

-- +migrate Down
DROP TABLE team_group;
