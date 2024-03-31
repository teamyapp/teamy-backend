-- +migrate Up
ALTER TABLE story ALTER COLUMN owner_id DROP NOT NULL;


-- +migrate Down
ALTER TABLE story ALTER COLUMN owner_id SET NOT NULL;
