-- +migrate Up
CREATE TABLE task_awaiting
(
	waiting_task_id BIGINT NOT NULL REFERENCES task (id) ON UPDATE CASCADE ON DELETE CASCADE,
	wait_for_task_id BIGINT NOT NULL REFERENCES task (id) ON UPDATE CASCADE ON DELETE CASCADE
);

-- +migrate Down
DROP TABLE task_awaiting;

