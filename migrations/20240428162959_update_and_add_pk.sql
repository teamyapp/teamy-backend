-- +migrate Up
ALTER TABLE app_package_upload_session
	ADD CONSTRAINT app_package_upload_session_pkey PRIMARY KEY (app_id, version_number, file_upload_session_id);
ALTER TABLE  apps_github_code_review
	DROP CONSTRAINT apps_github_code_review_pkey;
ALTER TABLE  apps_github_code_review
	DROP COLUMN github_reviewer_id;
ALTER TABLE  apps_github_code_review
	ADD CONSTRAINT apps_github_code_review_pkey PRIMARY KEY (github_pull_request_node_id, github_reviewer_node_id);
ALTER TABLE filter_group
	ADD CONSTRAINT filter_group_pkey PRIMARY KEY (group_id);
ALTER TABLE group_member_relation
	ADD CONSTRAINT group_member_relation_pkey PRIMARY KEY (group_id, member_id);
ALTER TABLE max_viewers_activator
	ADD CONSTRAINT max_viewers_activator_pkey PRIMARY KEY (activator_id);
ALTER TABLE percentage_activator
	ADD CONSTRAINT percentage_activator_pkey PRIMARY KEY (activator_id);
ALTER TABLE sprint_participant
	ADD CONSTRAINT sprint_participant_pkey PRIMARY KEY (sprint_id, user_id);
ALTER TABLE sprint_task_relation
	ADD CONSTRAINT sprint_task_relation_pkey PRIMARY KEY (sprint_id, task_id);
ALTER TABLE team_file_upload_session
	ADD CONSTRAINT team_file_upload_session_pkey PRIMARY KEY (team_id, type, file_upload_session_id);
ALTER TABLE time_range_activator
	ADD CONSTRAINT time_range_activator_pkey PRIMARY KEY (activator_id);
ALTER TABLE user_file_upload_session
	ADD CONSTRAINT user_file_upload_session_pkey PRIMARY KEY (user_id, type, file_upload_session_id);

-- +migrate Down
ALTER TABLE user_file_upload_session
	DROP CONSTRAINT user_file_upload_session_pkey;
ALTER TABLE time_range_activator
	DROP CONSTRAINT time_range_activator_pkey;
ALTER TABLE team_file_upload_session
	DROP CONSTRAINT team_file_upload_session_pkey;
ALTER TABLE sprint_task_relation
	DROP CONSTRAINT sprint_task_relation_pkey;
ALTER TABLE sprint_participant
	DROP CONSTRAINT sprint_participant_pkey;
ALTER TABLE percentage_activator
	DROP CONSTRAINT percentage_activator_pkey;
ALTER TABLE max_viewers_activator
	DROP CONSTRAINT max_viewers_activator_pkey;
ALTER TABLE group_member_relation
	DROP CONSTRAINT group_member_relation_pkey;
ALTER TABLE filter_group
	DROP CONSTRAINT filter_group_pkey;
ALTER TABLE  apps_github_code_review
	DROP CONSTRAINT apps_github_code_review_pkey;
ALTER TABLE  apps_github_code_review
	ADD COLUMN github_reviewer_id BIGINT NOT NULL DEFAULT 0;
ALTER TABLE  apps_github_code_review
	ADD CONSTRAINT apps_github_code_review_pkey PRIMARY KEY (github_pull_request_node_id, github_reviewer_id);
ALTER TABLE app_package_upload_session
	DROP CONSTRAINT app_package_upload_session_pkey;
