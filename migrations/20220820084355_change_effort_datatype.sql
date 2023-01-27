-- +migrate Up
ALTER TABLE task
    DROP COLUMN effort,
    ADD COLUMN effort BIGINT;

-- +migrate Down
ALTER TABLE task
    DROP COLUMN effort,
    ADD COLUMN effort INTEGER;
