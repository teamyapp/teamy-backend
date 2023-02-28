-- +migrate Up
ALTER TABLE apps_github_app_installation
	ADD github_organization_id BIGINT;
CREATE INDEX idx_github_organization_id
    ON apps_github_app_installation (github_organization_id);
CREATE INDEX idx_team_id
    ON apps_github_app_installation (team_id);
ALTER TABLE apps_github_app_installation
	DROP CONSTRAINT apps_github_app_installation_pkey;
ALTER TABLE apps_github_app_installation
	ADD PRIMARY KEY (id);

-- +migrate Down
ALTER TABLE apps_github_app_installation
	DROP COLUMN github_organization_id;
DROP INDEX idx_team_id;
ALTER TABLE apps_github_app_installation
	DROP CONSTRAINT apps_github_app_installation_pkey;
ALTER TABLE apps_github_app_installation
	ADD PRIMARY KEY (id, team_id);
