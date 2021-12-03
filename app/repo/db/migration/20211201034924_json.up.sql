CREATE TABLE json_persister (
   data json,
   id   INTEGER DEFAULT 1
);
insert into json_persister (data, id) values('{}', 1);