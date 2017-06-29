ALTER TABLE ONLY profiles
    ADD COLUMN title text;

ALTER TABLE ONLY profiles
	ADD COLUMN profile_type char(1);

ALTER TABLE ONLY profiles
	ADD COLUMN heading text;