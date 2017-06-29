ALTER TABLE ONLY profiles
	ADD COLUMN oauth_token text;

ALTER TABLE ONLY profiles
    ADD COLUMN oauth_expiry smallint;