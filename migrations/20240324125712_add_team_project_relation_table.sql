-- +migrate Up
CREATE TABLE team_project_relation (
    team_id BIGINT NOT NULL REFERENCES team (id) on DELETE CASCADE,
    project_id BIGINT NOT NULL REFERENCES project (id) on DELETE CASCADE,
    CONSTRAINT pk_team_project_relation PRIMARY KEY (team_id, project_id)
);

-- +migrate Down
DROP TABLE team_project_relation;

