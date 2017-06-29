CREATE SEQUENCE payment_history_payment_history_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
CREATE TABLE payment_history(
    payment_history_id  bigint DEFAULT nextval('payment_history_payment_history_id_seq'::regclass)  NOT NULL,
    payment_id bigint,
    amount      integer,
    status      text,
    message     text,
    user_id     bigint,   
    created     timestamp without time zone default (now() at time zone 'utc')
);

ALTER TABLE ONLY payment_history
    ADD CONSTRAINT payment_history_pkey PRIMARY KEY (payment_history_id);
ALTER TABLE ONLY payment_history
    ADD CONSTRAINT payment_history_payment_id_foreign_key FOREIGN KEY (payment_id) REFERENCES payment_requests(payment_request_id);