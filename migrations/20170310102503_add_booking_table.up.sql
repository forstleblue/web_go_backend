CREATE SEQUENCE     bookings_booking_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
CREATE TABLE bookings(
    booking_id  bigint DEFAULT nextval('bookings_booking_id_seq'::regclass) NOT NULL,
    profile_id bigint,
    user_id    bigint    
);


ALTER TABLE ONLY bookings
    ADD CONSTRAINT   bookings_pkey   PRIMARY KEY (booking_id);

ALTER TABLE ONLY bookings
    ADD CONSTRAINT   bookings_user_id_foreign_key FOREIGN KEY (user_id) REFERENCES users(user_id);

ALTER TABLE ONLY bookings
    ADD CONSTRAINT   bookings_profile_id_foreign_key FOREIGN KEY (profile_id) REFERENCES profiles(profile_id);
