-- +migrate Up
CREATE TABLE attachment_list (
    owner_type VARCHAR(255) NOT NULL,
    owner_id BIGINT NOT NULL,
    list_label VARCHAR(255) NOT NULL,
    list_id BIGINT NOT NULL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP
);

CREATE TABLE image (
    id BIGINT NOT NULL PRIMARY KEY,
    url VARCHAR(255) NOT NULL,
    size BIGINT NOT NULL,
    attachment_list_id BIGINT NOT NULL REFERENCES attachment_list (list_id),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP
);

CREATE TABLE task_file_upload_session (
    task_id BIGINT NOT NULL REFERENCES task (id),
    type VARCHAR(255) NOT NULL,
    file_upload_session_id BIGINT NOT NULL,
    is_completed BOOLEAN NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP,
    PRIMARY KEY (task_id, type, file_upload_session_id)
);

-- +migrate Down

DROP TABLE task_file_upload_session;
DROP TABLE image;
DROP TABLE attachment_list;