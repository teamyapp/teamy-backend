-- +migrate Up
ALTER TABLE filter_group
    DROP COLUMN count;

ALTER TABLE app_version
    DROP CONSTRAINT IF EXISTS app_version_created_by_user_id_fkey;

ALTER TABLE app_version
    ADD CONSTRAINT app_version_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) 
        REFERENCES "user" (id) ON DELETE SET NULL;

-- +migrate Down

ALTER TABLE app_version
    DROP CONSTRAINT IF EXISTS app_version_created_by_user_id_fkey;

ALTER TABLE app_version
ADD CONSTRAINT app_version_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) 
    REFERENCES "user" (id) ON DELETE CASCADE;

ALTER TABLE filter_group
    ADD count INT NOT NULL DEFAULT 0;
