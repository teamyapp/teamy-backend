-- +migrate Up
ALTER TABLE story ALTER COLUMN owner_id DROP NOT NULL;
ALTER TABLE story ADD COLUMN is_planned BOOLEAN DEFAULT FALSE;

-- +migrate Down
ALTER TABLE story DROP COLUMN is_planned;
ALTER TABLE story ALTER COLUMN owner_id SET NOT NULL;
