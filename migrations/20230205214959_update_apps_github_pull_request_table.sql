-- +migrate Up
ALTER TABLE apps_github_pull_request
	ADD github_pull_request_number integer,
    ADD github_pull_request_url VARCHAR(2048),
    ADD github_repository_name VARCHAR(20),
    ADD github_repository_owner VARCHAR(20);

-- +migrate Down
ALTER TABLE apps_github_pull_request
DROP COLUMN github_pull_request_number,
	DROP COLUMN github_pull_request_url,
	DROP COLUMN github_repository_name,
	DROP COLUMN github_repository_owner;
