CREATE SEQUENCE     servicecategory_service_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE servicecategory(
	service_id  bigint DEFAULT nextval('bookings_booking_id_seq'::regclass) NOT NULL,
    service_name    text,
    service_style_key   bigint
);

ALTER TABLE ONLY servicecategory
    ADD CONSTRAINT servicecategory_pkey PRIMARY KEY (service_id);

ALTER TABLE ONLY servicecategory 
    ADD CONSTRAINT  servicecategory_service_style_key_foreign_key 
    FOREIGN KEY (service_style_key) REFERENCES serviceinputtype(service_input_id);