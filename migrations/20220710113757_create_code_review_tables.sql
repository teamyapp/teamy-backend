-- +migrate Up
CREATE TABLE apps_github_pull_request
(
	internal_task_id BIGINT NOT NULL,
	github_pull_request_node_id VARCHAR(20) NOT NULL PRIMARY KEY
);

CREATE TABLE apps_github_code_review
(
    github_pull_request_node_id VARCHAR(20) NOT NULL,
    github_reviewer_id BIGINT NOT NULL,
	internal_code_review_task_id BIGINT NOT NULL,
	internal_address_feedback_task_id BIGINT,
	round SMALLINT NOT NULL,
	PRIMARY KEY (github_pull_request_node_id, github_reviewer_id, round)
);

-- +migrate Down
DROP TABLE apps_github_code_review;
DROP TABLE apps_github_pull_request;
