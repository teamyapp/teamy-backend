-- +migrate Up
CREATE TABLE task_link
(
    id   VARCHAR(50) NOT NULL,
	task_id BIGINT NOT NULL REFERENCES task (id) ON UPDATE CASCADE ON DELETE CASCADE,
    title TEXT NOT NULL,
	url VARCHAR(200) NOT NULL,
    icon_url TEXT,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP
);

-- +migrate Down
DROP TABLE task_link;
