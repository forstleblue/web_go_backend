ALTER TABLE ONLY widgets
	ADD COLUMN owner_type text;

ALTER TABLE ONLY widgets
	DROP CONSTRAINT widget_profile_id_foreign_key;

ALTER TABLE ONLY widgets
	RENAME COLUMN profile_id TO owner_id;

ALTER TABLE ONLY widgets
	ALTER COLUMN owner_id TYPE text;