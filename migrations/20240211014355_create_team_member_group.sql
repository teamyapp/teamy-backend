-- +migrate Up
ALTER TABLE team_group
	RENAME TO team_member_group;
ALTER TABLE team_member_group
	DROP CONSTRAINT team_group_pkey;
ALTER TABLE team_member_group
	ADD COLUMN id BIGINT PRIMARY KEY;
ALTER TABLE team_member_group
	RENAME COLUMN label TO name;
ALTER TABLE team_member_group
	ALTER COLUMN name TYPE VARCHAR(255);
ALTER TABLE team_member_group
	RENAME COLUMN user_group_id TO authorization_user_group_id;
ALTER TABLE team_member_group
	ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

CREATE TABLE team_member_group_user_relation
(
	group_id       BIGINT    NOT NULL REFERENCES team_member_group (id),
	member_user_id BIGINT    NOT NULL REFERENCES "user"(id),
	created_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	CONSTRAINT team_member_group_member_relation_pkey PRIMARY KEY (group_id, member_user_id)
);

DROP TABLE IF EXISTS team_group_relation;

-- +migrate Down
ALTER TABLE team_member_group_user_relation
	DROP CONSTRAINT team_member_group_member_relation_pkey;
DROP TABLE team_member_group_user_relation;

ALTER TABLE team_member_group
	DROP COLUMN updated_at;
ALTER TABLE team_member_group
	RENAME COLUMN authorization_user_group_id TO user_group_id;
ALTER TABLE team_member_group
	ALTER COLUMN name TYPE VARCHAR(20);
ALTER TABLE team_member_group
	RENAME COLUMN name TO label;
ALTER TABLE team_member_group
	DROP CONSTRAINT team_member_group_pkey;
ALTER TABLE team_member_group
	DROP COLUMN id;
ALTER TABLE team_member_group
	RENAME TO team_group;
ALTER TABLE team_group
	ADD PRIMARY KEY (team_id, user_group_id);
