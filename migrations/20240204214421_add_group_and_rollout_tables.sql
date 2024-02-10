-- +migrate Up
CREATE TABLE version_selector (
    id BIGINT PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE version_selector_version_relation (
    version_selector_id BIGINT NOT NULL REFERENCES version_selector (id) on DELETE CASCADE,
    version_number INT NOT NULL,
    CONSTRAINT pk_version_selector_version_relation PRIMARY KEY (version_selector_id, version_number)
);


CREATE TABLE activator (
    id BIGINT PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP
);

CREATE TABLE time_range_activator (
    activator_id BIGINT NOT NULL REFERENCES activator (id) on DELETE CASCADE,
    start_time TIME,
    end_time TIME
);

CREATE TABLE max_viewers_activator (
    activator_id BIGINT NOT NULL REFERENCES activator (id) on DELETE CASCADE,
    max_viewers INT NOT NULL
);

CREATE TABLE percentage_activator (
    activator_id BIGINT NOT NULL REFERENCES activator (id) on DELETE CASCADE,
    percentage INT NOT NULL
);

CREATE TABLE team_app_installation (
    id BIGINT PRIMARY KEY,
    installed_team_id BIGINT NOT NULL REFERENCES team (id) on DELETE CASCADE,
    app_id BIGINT NOT NULL REFERENCES app (id) on DELETE CASCADE
);

DROP TABLE IF EXISTS app_team_installation;

CREATE TABLE app_secret (
    id BIGINT PRIMARY KEY,
    secret VARCHAR(150) NOT NULL,
    name VARCHAR(50) NOT NULL,
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    added_by_user_id BIGINT NOT NULL REFERENCES "user" (id),
    last_used_at TIMESTAMP,
    app_id BIGINT NOT NULL REFERENCES app (id) on DELETE CASCADE
);

ALTER TABLE app_version
    ADD description VARCHAR(50) NOT NULL DEFAULT '',
    ADD icon_url VARCHAR(2048),
    -- TODO: make created_by_user_id and app_id NOT NULL, and change on DELETE action
    ADD created_by_user_id BIGINT REFERENCES "user" (id) ON DELETE CASCADE;

CREATE TABLE app_version_change (
    id BIGINT PRIMARY KEY,
    app_id BIGINT NOT NULL REFERENCES app (id) on DELETE CASCADE,
    version_number INT NOT NULL,
    change VARCHAR(50) NOT NULL
);

CREATE TABLE app_version_price (
    app_id BIGINT NOT NULL REFERENCES app (id) on DELETE CASCADE,
    version_number INT NOT NULL,
    currency VARCHAR(10) NOT NULL,
    amount INT NOT NULL DEFAULT 0,
    CONSTRAINT pk_app_version_price PRIMARY KEY (app_id, version_number, currency)
);

CREATE TABLE "group" (
    id BIGINT PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    member_type VARCHAR(50) NOT NULL,
    name VARCHAR(50) NOT NULL,
    max_rollout_index INT NOT NULL DEFAULT -1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE app_group_relation (
    app_id BIGINT NOT NULL REFERENCES app (id) on DELETE CASCADE,
    group_id BIGINT NOT NULL REFERENCES "group" (id) on DELETE CASCADE,
    CONSTRAINT pk_app_group_relation PRIMARY KEY (app_id, group_id)
);


CREATE TABLE group_member_relation (
    group_id BIGINT NOT NULL REFERENCES "group" (id) on DELETE CASCADE,
    member_id BIGINT NOT NULL
);

CREATE TABLE filter_group (
    group_id BIGINT NOT NULL REFERENCES "group" (id) on DELETE CASCADE,
    filter VARCHAR(50) NOT NULL,
    count INT NOT NULL
);

CREATE TABLE rollout (
    id BIGINT PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    activator_id BIGINT REFERENCES activator (id) on DELETE SET NULL,
    version_selector_id BIGINT REFERENCES version_selector (id) on DELETE SET NULL,
    viewers INT NOT NULL,
    is_enabled BOOLEAN NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE app_rollout_relation (
    app_id BIGINT NOT NULL REFERENCES app (id) on DELETE CASCADE,
    rollout_id BIGINT NOT NULL REFERENCES rollout (id) on DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    CONSTRAINT pk_app_rollout_relation PRIMARY KEY (app_id, rollout_id)
);

CREATE TABLE group_rollout_relation (
    group_id BIGINT NOT NULL REFERENCES "group" (id) on DELETE CASCADE,
    rollout_id BIGINT NOT NULL REFERENCES rollout (id) on DELETE CASCADE,
    order_index INT NOT NULL,
    CONSTRAINT pk_group_rollout_relation PRIMARY KEY (group_id, rollout_id)
);

CREATE TABLE rollout_viewer (
    rollout_id BIGINT NOT NULL REFERENCES rollout (id) on DELETE CASCADE,
    viewer_id BIGINT NOT NULL,
    version_number INT NOT NULL,
    is_activated BOOLEAN NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    CONSTRAINT pk_rollout_viewer PRIMARY KEY (rollout_id, viewer_id, version_number)
);

CREATE TABLE tag (
    id BIGINT PRIMARY KEY,
    value VARCHAR(50) NOT NULL UNIQUE
);

CREATE TABLE app_tag_relation (
    app_id BIGINT NOT NULL REFERENCES app (id) on DELETE CASCADE,
    tag_id BIGINT NOT NULL REFERENCES tag (id) on DELETE CASCADE,
    CONSTRAINT pk_app_tag_relation PRIMARY KEY (app_id, tag_id)
);

-- +migrate Down
DROP TABLE IF EXISTS app_tag_relation;
DROP TABLE IF EXISTS tag;
DROP TABLE IF EXISTS rollout_viewer;
DROP TABLE IF EXISTS group_rollout_relation;
DROP TABLE IF EXISTS app_rollout_relation;
DROP TABLE IF EXISTS rollout;
DROP TABLE IF EXISTS filter_group;
DROP TABLE IF EXISTS group_member_relation;
DROP TABLE IF EXISTS app_group_relation;
DROP TABLE IF EXISTS "group";
DROP TABLE IF EXISTS app_version_price;
DROP TABLE IF EXISTS app_version_change;
ALTER TABLE app_version
    DROP COLUMN description,
    DROP COLUMN icon_url,
    DROP COLUMN created_by_user_id;
DROP TABLE IF EXISTS app_secret;
CREATE TABLE IF NOT EXISTS app_team_installation (
	app_id BIGINT NOT NULL REFERENCES app (id) 
	    ON UPDATE CASCADE 
	    ON DELETE CASCADE,
	installed_team_id BIGINT NOT NULL REFERENCES team (id) 
	    ON UPDATE CASCADE 
	    ON DELETE CASCADE,
	installed_by_user_id BIGINT NOT NULL REFERENCES "user" (id) 
	    ON UPDATE CASCADE 
	    ON DELETE SET NULL,
	enabled_version_number INT NOT NULL,
	installed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (app_id, installed_team_id)
);
DROP TABLE IF EXISTS team_app_installation;
DROP TABLE IF EXISTS percentage_activator;
DROP TABLE IF EXISTS max_viewers_activator;
DROP TABLE IF EXISTS time_range_activator;
DROP TABLE IF EXISTS activator;
DROP TABLE IF EXISTS version_selector_version_relation;
DROP TABLE IF EXISTS version_selector;
