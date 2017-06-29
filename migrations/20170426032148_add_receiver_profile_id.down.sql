ALTER TABLE ONLY feedbacks
	DROP COLUMN IF EXISTS receiver_profile_id;

ALTER TABLE ONLY feedbacks
	DROP COLUMN IF EXISTS created;
    
ALTER TABLE ONLY feedbacks
    RENAME sender_profile_id TO created_profile_id;