-- +migrate Up
CREATE INDEX index_attachment_list_on_owner_type_and_owner_id_and_list_label ON attachment_list (owner_type, owner_id, list_label);

-- +migrate Down
DROP INDEX index_attachment_list_on_owner_type_and_owner_id_and_list_label;
