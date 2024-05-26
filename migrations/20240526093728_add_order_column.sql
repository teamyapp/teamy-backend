-- +migrate Up
ALTER TABLE team_member_group ADD COLUMN "order" INT NOT NULL DEFAULT 0;
-- +migrate Down
ALTER TABLE team_member_group DROP COLUMN "order";

