-- +migrate Up
CREATE TABLE project (
    id BIGINT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    expected_start_at TIMESTAMP,
    actual_start_at TIMESTAMP,
    expected_end_at TIMESTAMP,
    actual_end_at TIMESTAMP,
    creator_id BIGINT NOT NULL REFERENCES "user"(id),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP
);

CREATE TABLE phase (
    id BIGINT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL,
    expected_start_at TIMESTAMP NOT NULL,
    actual_start_at TIMESTAMP,
    expected_end_at TIMESTAMP NOT NULL,
    actual_end_at TIMESTAMP,
    creator_id BIGINT NOT NULL REFERENCES "user"(id),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP
);

CREATE TABLE story (
    id BIGINT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    owner_id BIGINT NOT NULL REFERENCES "user"(id),
    status VARCHAR(50) NOT NULL,
    priority VARCHAR(50),
    creator_id BIGINT NOT NULL REFERENCES "user"(id),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP
);

CREATE TABLE project_story_relation (
    project_id BIGINT NOT NULL REFERENCES project(id),
    story_id BIGINT NOT NULL REFERENCES story(id),
    PRIMARY KEY (project_id, story_id)
);

CREATE TABLE project_phase_relation (
    project_id BIGINT NOT NULL REFERENCES project(id),
    phase_id BIGINT NOT NULL REFERENCES phase(id),
    PRIMARY KEY (project_id, phase_id)
);

CREATE TABLE phase_story_relation (
    phase_id BIGINT NOT NULL REFERENCES phase(id),
    story_id BIGINT NOT NULL REFERENCES story(id),
    PRIMARY KEY (phase_id, story_id)
);

CREATE TABLE story_task_relation (
    story_id BIGINT NOT NULL REFERENCES story(id),
    task_id BIGINT NOT NULL REFERENCES task(id),
    PRIMARY KEY (story_id, task_id)
);

-- +migrate Down
DROP TABLE story_task_relation;
DROP TABLE phase_story_relation;
DROP TABLE project_phase_relation;
DROP TABLE project_story_relation;
DROP TABLE story;
DROP TABLE phase;
DROP TABLE project;

