ALTER TABLE ONLY widgets
	ALTER COLUMN owner_id TYPE uuid USING(owner_id::uuid);