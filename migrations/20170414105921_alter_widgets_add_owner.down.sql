ALTER TABLE ONLY widgets
	DROP COLUMN IF EXISTS owner_id;

ALTER TABLE ONLY widgets
	ADD COLUMN profile_id bigint;

ALTER TABLE ONLY widgets
    ADD CONSTRAINT   widget_profile_id_foreign_key FOREIGN KEY (profile_id) REFERENCES profiles(profile_id);

ALTER TABLE ONLY widgets
	DROP COLUMN IF EXISTS owner_type;