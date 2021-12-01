CREATE TABLE "user"
(
	id          SERIAL PRIMARY KEY,
	first_name  VARCHAR(100) NOT NULL,
	last_name   VARCHAR(100) NOT NULL,
	profile_url VARCHAR(2048),
	created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at  TIMESTAMP
);

CREATE TABLE task
(
	id               SERIAL PRIMARY KEY,
	goal             VARCHAR(240) NOT NULL,
	due_at           TIMESTAMP,
	context          TEXT,
	owner_user_id    INTEGER REFERENCES "user" ON UPDATE CASCADE,
	work_scope_index INTEGER   DEFAULT 0,
	effort           INTEGER,
	num_of_unknowns  INTEGER,
	created_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at       TIMESTAMP
);

CREATE TABLE task_dependency
(
	need_before INTEGER NOT NULL REFERENCES task ON UPDATE CASCADE,
	need_after  INTEGER NOT NULL REFERENCES task ON UPDATE CASCADE,
	CONSTRAINT pk_task_dependency PRIMARY KEY (need_before, need_after)
);

CREATE TABLE team
(
	id         SERIAL PRIMARY KEY,
	name       VARCHAR(50) NOT NULL,
	logo_url   VARCHAR(2048),
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP
);

CREATE TABLE team_member
(
	team_id                INTEGER NOT NULL REFERENCES team ON UPDATE CASCADE,
	user_id                INTEGER NOT NULL REFERENCES "user" ON UPDATE CASCADE,
	need_attention_task_id INTEGER REFERENCES task ON UPDATE CASCADE,
	created_at             TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at             TIMESTAMP,
	CONSTRAINT pk_team_members PRIMARY KEY (team_id, user_id)
);


CREATE TABLE task_status
(
	value SMALLINT PRIMARY KEY,
	name  VARCHAR(50)
);

CREATE TABLE team_task
(
	team_id     INTEGER NOT NULL REFERENCES team ON UPDATE CASCADE,
	task_id     INTEGER NOT NULL REFERENCES task ON UPDATE CASCADE,
	task_status SMALLINT REFERENCES task_status ON UPDATE CASCADE
);

CREATE TABLE user_state
(
	user_id        INTEGER REFERENCES "user" ON UPDATE CASCADE,
	active_team_id INTEGER REFERENCES team ON UPDATE CASCADE
);
