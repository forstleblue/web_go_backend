ALTER TABLE ONLY payment_requests
DROP COLUMN message;

ALTER TABLE ONLY payment_requests
ADD COLUMN message jsonb;
CREATE INDEX idx_gin_message ON payment_requests USING gin (message);