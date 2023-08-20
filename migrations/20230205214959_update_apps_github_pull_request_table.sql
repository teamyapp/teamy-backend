-- +migrate Up
ALTER TABLE apps_github_pull_request
	ADD github_repository_owner VARCHAR(39),
	ADD github_repository_name VARCHAR(100),
    ADD github_pull_request_number INTEGER,
    ADD github_pull_request_url VARCHAR(2048);

-- +migrate Down
ALTER TABLE apps_github_pull_request
	DROP COLUMN github_repository_owner,
	DROP COLUMN github_repository_name,
    DROP COLUMN github_pull_request_number,
	DROP COLUMN github_pull_request_url;
