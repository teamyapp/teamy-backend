-- +migrate Up
ALTER TABLE apps_github_required_user_action
	DROP COLUMN id;
ALTER TABLE apps_github_required_user_action
	ADD CONSTRAINT apps_github_required_user_action_pkey PRIMARY KEY (team_id, user_action_type, action_user_id);

-- +migrate Down
ALTER TABLE apps_github_required_user_action
	DROP CONSTRAINT apps_github_required_user_action_pkey;
ALTER TABLE apps_github_required_user_action
	ADD COLUMN id BIGINT PRIMARY KEY DEFAULT 0;
