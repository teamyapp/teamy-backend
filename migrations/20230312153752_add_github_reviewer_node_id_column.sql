-- +migrate Up
ALTER TABLE apps_github_code_review
	ADD github_reviewer_node_id VARCHAR(20);

-- +migrate Down
ALTER TABLE apps_github_code_review
	DROP github_reviewer_node_id;
