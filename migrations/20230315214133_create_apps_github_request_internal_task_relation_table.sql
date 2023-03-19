-- +migrate Up
CREATE TABLE apps_github_pull_request_internal_task_relation (
	internal_task_id BIGINT NOT NULL,
	pull_request_node_id VARCHAR(20) NOT NULL REFERENCES apps_github_pull_request (github_pull_request_node_id) ON UPDATE CASCADE ON DELETE CASCADE,
	automatic_tracking BOOL,
	pull_request_task_link_id BIGINT NOT NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	CONSTRAINT pk_apps_github_pull_request_internal_task_relation PRIMARY KEY (pull_request_node_id, internal_task_id)
);

CREATE INDEX idx_apps_github_pull_request_internal_task_relation_internal_task_id
    ON apps_github_pull_request_internal_task_relation(internal_task_id);

ALTER TABLE
	apps_github_pull_request DROP internal_task_id;

-- +migrate Down
ALTER TABLE
	apps_github_pull_request
ADD
	internal_task_id BIGINT;

DROP INDEX idx_apps_github_pull_request_internal_task_relation_internal_task_id DROP TABLE apps_github_pull_request_internal_task_relation;