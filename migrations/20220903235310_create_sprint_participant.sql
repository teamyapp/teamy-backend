-- +migrate Up
CREATE TABLE sprint_participant (
    sprint_id BIGINT NOT NULL REFERENCES sprint(id) ON UPDATE CASCADE ON DELETE CASCADE,
	user_id BIGINT NOT NULL REFERENCES "user"(id) ON UPDATE CASCADE ON DELETE CASCADE,
	total_bandwidth BIGINT NOT NULL DEFAULT 0,
	unused_bandwidth BIGINT NOT NULL DEFAULT 0,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP
);

-- +migrate Down
DROP TABLE sprint_participant;
