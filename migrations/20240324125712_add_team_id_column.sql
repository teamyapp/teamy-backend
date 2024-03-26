-- +migrate Up
ALTER TABLE project ADD COLUMN team_id BIGINT NOT NULL REFERENCES team(id) ON DELETE CASCADE;

-- +migrate Down
ALTER TABLE project DROP COLUMN team_id;