CREATE SEQUENCE     serviceinputtype_service_input_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE serviceinputtype(
    service_input_id  bigint DEFAULT nextval('bookings_booking_id_seq'::regclass) NOT NULL,
    from_date   boolean,
    to_date     boolean,
    from_time   boolean,
    to_time     boolean,
    frequency_unit  boolean
);


ALTER TABLE ONLY serviceinputtype
    ADD CONSTRAINT   serviceinputtype_pkey   PRIMARY KEY (service_input_id);