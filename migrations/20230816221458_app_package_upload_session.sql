-- +migrate Up
CREATE TABLE app_package_upload_session (
	app_id BIGINT NOT NULL REFERENCES app (id),
    user_id BIGINT NOT NULL REFERENCES "user" (id),
    file_upload_session_id BIGINT NOT NULL,
	version_number INT NOT NULL,
	is_completed BOOLEAN NOT NULL DEFAULT FALSE,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP
);

-- +migrate Down
DROP TABLE app_package_upload_session;
