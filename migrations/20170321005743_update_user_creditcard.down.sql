ALTER TABLE ONLY users
    RENAME credit_card_id TO credit_card;
ALTER TABLE ONLY users
    DROP COLUMN IF EXISTS credit_card_mask;