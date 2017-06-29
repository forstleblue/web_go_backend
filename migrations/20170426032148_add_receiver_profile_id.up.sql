ALTER TABLE ONLY feedbacks
	ADD COLUMN receiver_profile_id bigint;

ALTER TABLE ONLY feedbacks
	RENAME created_profile_id TO sender_profile_id;

ALTER TABLE ONLY feedbacks
	ADD COLUMN created timestamp without time zone default (now() at time zone 'utc')