-- +migrate Up
ALTER table project DROP COLUMN iconURL;
ALTER table project ADD COLUMN icon_url VARCHAR(255);

-- +migrate Down
ALTER table project DROP COLUMN icon_url;
ALTER table project ADD COLUMN iconURL VARCHAR(255);
