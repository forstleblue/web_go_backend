ALTER TABLE ONLY messages
    DROP CONSTRAINT messages_profile_id_foreign_key;

ALTER TABLE ONLY messages
    RENAME  profile_id TO user_id;  

