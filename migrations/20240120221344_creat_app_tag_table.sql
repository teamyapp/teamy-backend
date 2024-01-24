-- +migrate Up
CREATE TABLE tag (
    id BIGINT PRIMARY KEY,
    "value" VARCHAR(50) NOT NULL UNIQUE
);

CREATE TABLE app_tag_relation (
    app_id BIGINT NOT NULL REFERENCES app (id) on DELETE CASCADE,
    tag_id BIGINT NOT NULL REFERENCES tag (id) on DELETE CASCADE,
    CONSTRAINT pk_app_tag_relation PRIMARY KEY (app_id, tag_id)
);

-- +migrate Down
DROP TABLE app_tag_relation;
DROP TABLE tag;
