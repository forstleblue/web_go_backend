ALTER TABLE ONLY messages
    DROP CONSTRAINT messages_user_id_foreign_key;

ALTER TABLE ONLY messages
    RENAME user_id TO profile_id;

ALTER TABLE ONLY messages
    ADD CONSTRAINT messages_profile_id_foreign_key FOREIGN KEY (profile_id) REFERENCES profiles(profile_id);