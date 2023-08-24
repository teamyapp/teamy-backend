-- +migrate Up
ALTER TABLE app
	RENAME COLUMN installation_count TO total_installations;
ALTER TABLE app
	ADD managed_by_team_id BIGINT NOT NULL REFERENCES "team" (id) ON UPDATE CASCADE ON DELETE CASCADE;
ALTER TABLE app
	DROP COLUMN api_secret;
ALTER TABLE app
	DROP COLUMN active_version_number;
ALTER TABLE app
	DROP COLUMN creator_user_id;
ALTER TABLE app
	DROP COLUMN "name";
ALTER TABLE app
	DROP COLUMN "description";

ALTER TABLE app_version 
	RENAME COLUMN is_public TO is_ready;
ALTER TABLE app_version
	RENAME COLUMN version_number TO "number";
ALTER TABLE app_version
	ADD app_name VARCHAR(100) NOT NULL;
ALTER TABLE app_version
	DROP COLUMN icon_url;
ALTER TABLE app_version
	DROP COLUMN has_ui_extension;
ALTER TABLE app_version
	DROP COLUMN ui_extension_entrypoint_path;
ALTER TABLE app_version
	DROP COLUMN "changes";

-- +migrate Down
ALTER TABLE app_version
	ADD "changes" TEXT;
ALTER TABLE app_version
	ADD ui_extension_entrypoint_path VARCHAR(4096);
ALTER TABLE app_version
	ADD has_ui_extension BOOL NOT NULL DEFAULT FALSE;
ALTER TABLE app_version
	ADD icon_url VARCHAR(2048);
ALTER TABLE app_version
	DROP COLUMN app_name;
ALTER TABLE app_version
	RENAME COLUMN "number" TO version_number;
ALTER TABLE app_version 
	RENAME COLUMN is_ready TO is_public;

ALTER TABLE app
	ADD "description" TEXT NOT NULL;
ALTER TABLE app
	ADD "name" VARCHAR(100) NOT NULL;
ALTER TABLE app
	ADD creator_user_id BIGINT NOT NULL REFERENCES "user" (id) ON UPDATE CASCADE ON DELETE CASCADE;
ALTER TABLE app
	ADD active_version_number INT;
ALTER TABLE app
	ADD api_secret VARCHAR(128) NOT NULL;
ALTER TABLE app
	DROP COLUMN managed_by_team_id;
ALTER TABLE app
	RENAME COLUMN total_installations TO installation_count;






