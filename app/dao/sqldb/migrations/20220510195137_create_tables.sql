-- +migrate Up
CREATE TABLE "user"
(
    id          BIGINT PRIMARY KEY,
    first_name  VARCHAR(50) NOT NULL,
	last_name   VARCHAR(50) NOT NULL,
	profile_url VARCHAR(2048),
	created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at  TIMESTAMP
);

CREATE TABLE team
(
	id         BIGINT PRIMARY KEY,
	name       VARCHAR(50) NOT NULL,
	icon_url   VARCHAR(2048),
	creator_id BIGINT NOT NULL REFERENCES "user"(id) ON UPDATE CASCADE,
	owner_id   BIGINT NOT NULL REFERENCES "user"(id) ON UPDATE CASCADE,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP
);

CREATE TABLE thread
(
	id BIGINT PRIMARY KEY
);

CREATE TABLE task
(
	id               BIGINT PRIMARY KEY,
	created_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at       TIMESTAMP,
	goal             VARCHAR(240) NOT NULL,
	due_at           TIMESTAMP,
	context          TEXT,
	creator_user_id  BIGINT NOT NULL REFERENCES "user"(id) ON UPDATE CASCADE,
	owner_user_id    BIGINT REFERENCES "user"(id) ON UPDATE CASCADE,
	owning_team_id	 BIGINT NOT NULL REFERENCES team(id) ON UPDATE CASCADE,
	status           VARCHAR(20) NOT NULL,
	effort			 INTEGER,
	comments_thread_id BIGINT NOT NULL REFERENCES thread(id) ON UPDATE CASCADE
);

CREATE TABLE team_member
(
	team_id                BIGINT NOT NULL REFERENCES team(id) ON UPDATE CASCADE,
	user_id                BIGINT NOT NULL REFERENCES "user"(id) ON UPDATE CASCADE,
	need_attention_task_id BIGINT REFERENCES task(id) ON UPDATE CASCADE,
	created_at             TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at             TIMESTAMP,
	CONSTRAINT pk_team_member PRIMARY KEY (team_id, user_id)
);

CREATE TABLE message
(
	id BIGINT PRIMARY KEY,
	body TEXT NOT NULL,
	thread_id BIGINT NOT NULL REFERENCES thread(id) ON UPDATE CASCADE,
	author_user_id BIGINT NOT NULL REFERENCES "user"(id) ON UPDATE CASCADE,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP
);

CREATE TABLE invitation
(
	id BIGINT PRIMARY KEY,
	sender_user_id BIGINT NOT NULL REFERENCES "user"(id) ON UPDATE CASCADE,
	receiver_first_name VARCHAR(50),
	receiver_last_name VARCHAR(50),
	receiver_user_id BIGINT REFERENCES "user"(id) ON UPDATE CASCADE,
	receiver_email VARCHAR(100),
	team_id BIGINT NOT NULL REFERENCES team(id) ON UPDATE CASCADE,
	expire_at TIMESTAMP NOT NULL,
	status VARCHAR(20) NOT NULL,
	code VARCHAR(50) NOT NULL,
	create_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- +migrate Down
DROP TABLE invitation;
DROP TABLE message;
DROP TABLE team_member;
DROP TABLE task;
DROP TABLE thread;
DROP TABLE team;
DROP TABLE "user";

