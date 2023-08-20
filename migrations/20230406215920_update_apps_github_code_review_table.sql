-- +migrate Up
ALTER TABLE apps_github_code_review
	ADD CONSTRAINT fk_github_pull_request_node_id
    	FOREIGN KEY (github_pull_request_node_id) REFERENCES apps_github_pull_request(github_pull_request_node_id);

-- +migrate Down
ALTER TABLE apps_github_code_review
	DROP CONSTRAINT fk_github_pull_request_node_id;
