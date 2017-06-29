CREATE SEQUENCE feedbacks_feedback_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
CREATE TABLE feedbacks(
    feedback_id bigint DEFAULT nextval('feedbacks_feedback_id_seq'::regclass) NOT NULL,
    booking_id  bigint,
    created_profile_id bigint,
    description text,
    comment     text,
    score       smallint,
    sda_text    text[]
);

ALTER TABLE ONLY feedbacks  
    ADD CONSTRAINT feedbacks_pkey   PRIMARY KEY (feedback_id);
ALTER TABLE ONLY feedbacks
    ADD CONSTRAINT feedbacks_booking_id_foreign_key FOREIGN KEY (booking_id) REFERENCES bookings(booking_id);

