-- +migrate Up
CREATE TABLE team_member_group_invitation_relation (
	group_id       BIGINT    NOT NULL REFERENCES team_member_group (id) ON UPDATE CASCADE ON DELETE CASCADE,
	invitation_id BIGINT    NOT NULL REFERENCES invitation (id) ON UPDATE CASCADE ON DELETE CASCADE,
	created_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	CONSTRAINT team_member_group_invitation_relation_pkey PRIMARY KEY (group_id, invitation_id)
);

-- +migrate Down
DROP TABLE team_member_group_invitation_relation;
