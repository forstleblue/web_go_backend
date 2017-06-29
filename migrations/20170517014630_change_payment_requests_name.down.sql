ALTER TABLE ONLY payments
    RENAME COLUMN payment_id TO payment_request_id;
ALTER TABLE ONLY payments
    RENAME COLUMN payment_uuid TO payment_request_uuid;
ALTER TABLE ONLY payments
    RENAME TO payment_requests;