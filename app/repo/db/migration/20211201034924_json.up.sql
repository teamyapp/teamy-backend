CREATE TABLE json_persister (
   data json,
   id   INTEGER DEFAULT 1
);
INSERT INTO json_persister (data, id) VALUES ('{}', 1);