-- +migrate Up
ALTER TABLE apps_github_pull_request
	ADD github_organization_id BIGINT;

-- +migrate Down
ALTER TABLE apps_github_pull_request
	DROP COLUMN github_organization_id;
