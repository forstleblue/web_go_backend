CREATE SEQUENCE     payment_requests_payment_request_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
CREATE TABLE payment_requests(
    payment_request_id  bigint DEFAULT nextval('payment_requests_payment_request_id_seq'::regclass) NOT NULL,
    booking_id bigint,
    amount    integer,
    request_date  timestamp without time zone,
    confirmed_date  timestamp without time zone,
    payment_date  timestamp without time zone,
    payment_method text,
    transaction_id text,
    payment_status text,
    acct_display text,
    paypal_token text,
    paypal_payer_id text,
    paypal_payer_status text,
    status text
);


ALTER TABLE ONLY payment_requests
    ADD CONSTRAINT   payment_request_pkey   PRIMARY KEY (payment_request_id);

ALTER TABLE ONLY payment_requests
    ADD CONSTRAINT   payment_requests_booking_id_foreign_key FOREIGN KEY (booking_id) REFERENCES bookings(booking_id);