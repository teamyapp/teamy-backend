-- +migrate Up
ALTER TABLE team_member
	ADD weekly_bandwidth BIGINT NOT NULL DEFAULT 0;

-- +migrate Down
ALTER TABLE team_member
	DROP COLUMN weekly_bandwidth;
