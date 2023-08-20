-- +migrate Up
ALTER TABLE team
	ADD active_sprint_id BIGINT REFERENCES sprint (id) 
	    ON UPDATE CASCADE 
	    ON DELETE CASCADE;

-- +migrate Down
ALTER TABLE team
	DROP active_sprint_id;