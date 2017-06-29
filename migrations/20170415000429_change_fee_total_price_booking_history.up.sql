ALTER TABLE ONLY booking_history
    ALTER COLUMN total_price TYPE float8 USING CAST(total_price AS FLOAT);

ALTER TABLE ONLY booking_history
    ALTER COLUMN fee TYPE float8 USING CAST(fee AS FLOAT);