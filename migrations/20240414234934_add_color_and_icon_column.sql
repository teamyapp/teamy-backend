-- +migrate Up
ALTER table project ADD COLUMN color VARCHAR(50);
ALTER table project ADD COLUMN iconURL VARCHAR(255);

-- +migrate Down
ALTER table project DROP COLUMN color;
ALTER table project DROP COLUMN iconURL;
