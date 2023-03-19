-- +migrate Up
ALTER TABLE apps_github_code_review
    DROP CONSTRAINT apps_github_code_review_pkey,
	ADD github_reviewer_node_id VARCHAR(20) DEFAULT "",
    ADD PRIMARY KEY (github_pull_request_node_id, github_reviewer_node_id);

-- +migrate Down
ALTER TABLE apps_github_code_review
    DROP CONSTRAINT apps_github_code_review_pkey,
	DROP github_reviewer_node_id,
    ADD PRIMARY KEY (github_pull_request_node_id, github_reviewer_id);
