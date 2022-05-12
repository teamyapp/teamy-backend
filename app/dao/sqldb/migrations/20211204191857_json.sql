-- +migrate Up
CREATE TABLE json_persister
(
	data json,
	id   INTEGER DEFAULT 1
);
INSERT INTO json_persister (data, id)
VALUES ('{}', 1);

-- +migrate Down
DROP TABLE IF EXISTS
	json_persister;
