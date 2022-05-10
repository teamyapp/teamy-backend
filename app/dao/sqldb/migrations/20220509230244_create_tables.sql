-- +migrate Up
CREATE TABLE invitation
(
    id BIGINT,
	sender_user_id BIGINT,
	receiver_first_name VARCHAR(50),
	receiver_last_name VARCHAR(50),
	receiver_user_id BIGINT,
	receiver_email VARCHAR(100),
	team_id BIGINT,
	expire_at TIMESTAMP,
	status VARCHAR(20),
	code VARCHAR(50),
	create_at TIMESTAMP
);
-- +migrate Down
DROP TABLE invitation;
