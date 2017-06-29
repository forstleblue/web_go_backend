ALTER TABLE ONLY notifications
	RENAME COLUMN notification_string TO notification_text;
ALTER TABLE ONLY notifications
	ADD COLUMN entity_history_id bigint;
	