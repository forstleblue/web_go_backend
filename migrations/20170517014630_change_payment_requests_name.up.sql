ALTER TABLE ONLY payment_requests
    RENAME COLUMN payment_request_id TO payment_id;
ALTER TABLE ONLY payment_requests
    RENAME COLUMN payment_request_uuid TO payment_uuid;
ALTER TABLE ONLY payment_requests
    RENAME TO payments;