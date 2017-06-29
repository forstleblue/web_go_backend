CREATE SEQUENCE     booking_history_booking_history_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
CREATE TABLE booking_history(
    booking_history_id  bigint DEFAULT nextval('booking_history_booking_history_id_seq'::regclass) NOT NULL,
    booking_id              bigint,
    user_id                 bigint,
    message                 text,
    from_time               text,
    to_time                 text,
    from_date               text,
    to_date                 text,
    address                 text,
    fee                     integer,
    total_price             integer,
    booking_status          text,
    frequency_unit          text,
    frequency_value         bigint,
    created                 timestamp without time zone
);

ALTER TABLE ONLY booking_history
    ADD CONSTRAINT booking_history_pkey PRIMARY KEY (booking_history_id);

ALTER TABLE ONLY booking_history
    ADD CONSTRAINT   booking_history_booking_id_foreign_key FOREIGN KEY (booking_id) REFERENCES bookings(booking_id);
