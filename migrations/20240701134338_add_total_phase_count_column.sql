-- +migrate Up
ALTER TABLE project ADD COLUMN total_phase_count INT NOT NULL DEFAULT 0;
ALTER TABLE project ADD COLUMN completed_phase_count INT NOT NULL DEFAULT 0;

-- +migrate Down
ALTER TABLE project DROP COLUMN total_phase_count;
ALTER TABLE project DROP COLUMN completed_phase_count;

