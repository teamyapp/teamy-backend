-- +migrate Up
CREATE TABLE apps_github_required_user_action
(
    id BIGINT NOT NULL PRIMARY KEY,
    team_id BIGINT NOT NULL REFERENCES team (id) ON UPDATE CASCADE ON DELETE CASCADE,
    action_user_id BIGINT NOT NULL REFERENCES "user" (id) ON UPDATE CASCADE ON DELETE CASCADE,
    user_action_type VARCHAR(50) NOT NULL,
    is_completed BOOLEAN NOT NULL DEFAULT FALSE,
    requested_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    requested_by_user_id BIGINT NOT NULL REFERENCES "user" (id) ON UPDATE CASCADE ON DELETE CASCADE
);

-- +migrate Down
DROP TABLE apps_github_required_user_action;
