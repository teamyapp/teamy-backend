-- +migrate Up
CREATE TABLE user_file_upload_session
(
	user_id BIGINT NOT NULL REFERENCES "user" (id) ON UPDATE CASCADE ON DELETE CASCADE,
	type VARCHAR(50) NOT NULL,
	file_upload_session_id BIGINT NOT NULL,
	is_completed BOOLEAN NOT NULL DEFAULT FALSE,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP
);

CREATE TABLE team_file_upload_session
(
	team_id BIGINT NOT NULL REFERENCES team (id) ON UPDATE CASCADE ON DELETE CASCADE,
	type VARCHAR(50) NOT NULL,
	file_upload_session_id BIGINT NOT NULL,
	is_completed BOOLEAN NOT NULL DEFAULT FALSE,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP
);

-- +migrate Down
DROP TABLE team_file_upload_session;
DROP TABLE user_file_upload_session;
