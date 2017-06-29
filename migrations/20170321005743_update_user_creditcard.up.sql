ALTER TABLE ONLY users
    RENAME credit_card TO credit_card_id;
ALTER TABLE ONLY users
    ADD COLUMN credit_card_mask text;