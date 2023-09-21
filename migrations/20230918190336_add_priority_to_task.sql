-- +migrate Up
ALTER TABLE task
	ADD priority VARCHAR(20);

-- +migrate Down
ALTER TABLE task
	DROP priority;
