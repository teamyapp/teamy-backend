-- +migrate Up
CREATE TABLE group_member_relation (
    group_id BIGINT NOT NULL REFERENCES "group" (id) on DELETE CASCADE,
    member_id INTEGER NOT NULL
);

ALTER TABLE "group"
    ADD "member_type" VARCHAR(50) NOT NULL DEFAULT 'TEAM';

DROP TABLE user_group_relation;
DROP TABLE team_group_relation;
ALTER TABLE app_group_relation
    DROP "type";

-- +migrate Down
ALTER TABLE app_group_relation
    ADD "type" VARCHAR(50) NOT NULL DEFAULT 'TEAM';

CREATE TABLE team_group_relation (
    team_id BIGINT NOT NULL REFERENCES team (id) on DELETE CASCADE,
    group_id BIGINT NOT NULL REFERENCES "group" (id) on DELETE CASCADE,
    CONSTRAINT pk_team_group_relation PRIMARY KEY (team_id, group_id)
);

CREATE TABLE user_group_relation (
    user_id BIGINT NOT NULL REFERENCES "user" (id) on DELETE CASCADE,
    group_id BIGINT NOT NULL REFERENCES "group" (id) on DELETE CASCADE,
    CONSTRAINT pk_user_group_relation PRIMARY KEY (user_id, group_id)
);

ALTER TABLE "group"
    DROP "member_type";

DROP TABLE group_member_relation;
