ALTER TABLE ONLY booking_history
    ALTER COLUMN total_price TYPE numeric(12,2) USING CAST(total_price AS numeric);

ALTER TABLE ONLY booking_history
    ALTER COLUMN fee TYPE numeric(12,2) USING CAST(fee AS numeric);