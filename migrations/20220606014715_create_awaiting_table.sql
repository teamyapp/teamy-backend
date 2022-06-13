-- +migrate Up
CREATE TABLE task_await_for_relation
(
	awaiting_task_id BIGINT NOT NULL REFERENCES task (id) ON UPDATE CASCADE ON DELETE CASCADE,
	await_for_task_id BIGINT NOT NULL REFERENCES task (id) ON UPDATE CASCADE ON DELETE CASCADE,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (awaiting_task_id, await_for_task_id)
);

-- +migrate Down
DROP TABLE task_await_for_relation;
