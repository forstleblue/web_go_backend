CREATE SEQUENCE     platform_ebay_feedback_comment_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
CREATE TABLE platform_ebay(
    feedback_comment_id  bigint DEFAULT nextval('platform_ebay_feedback_comment_id_seq'::regclass) NOT NULL,
    profile_id      bigint,
    commenting_user  text,
    commenting_user_score text,
    comment_text text,
    comment_time timestamp without time zone,
    comment_type text,
    item_id text,
    role text,   
    feedback_id text,
    transaction_id text,
    item_title  text,
    item_price text
);

ALTER TABLE ONLY platform_ebay
    ADD CONSTRAINT   platform_ebay_pkey   PRIMARY KEY (feedback_comment_id);

ALTER TABLE ONLY platform_ebay
    ADD CONSTRAINT   platform_ebay_profile_id_foreign_key FOREIGN KEY (profile_id) REFERENCES profiles(profile_id);