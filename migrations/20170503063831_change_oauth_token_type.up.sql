ALTER TABLE ONLY profiles
	DROP COLUMN IF EXISTS oauth_expiry;
ALTER TABLE ONLY profiles
	ADD COLUMN oauth_expiry timestamp without time zone;