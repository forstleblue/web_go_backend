CREATE EXTENSION IF NOT EXISTS hstore;

CREATE SEQUENCE profiles_profile_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE profiles (
    profile_id bigint DEFAULT nextval('profiles_profile_id_seq'::regclass) NOT NULL,
    user_id bigint,
    skill text,
    tags text,
    photo_url text,
    description text,
    feedback_rating integer,
    reputation_status integer,
    fee text,
    created timestamp without time zone,
    updated timestamp without time zone
);

ALTER TABLE ONLY profiles
    ADD CONSTRAINT profiles_pkey PRIMARY KEY (profile_id);


CREATE SEQUENCE users_user_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE users (
    user_id bigint DEFAULT nextval('users_user_id_seq'::regclass) NOT NULL,
    first_name text,
    middle_name text,
    last_name text,
    email text,
    date_of_birth timestamp without time zone,
    facebook_id text,
    paypal_id text,
    credit_card text,
    password text,
    phones hstore,
    last_login hstore,
    created timestamp without time zone,
    updated timestamp without time zone,
    photo_url text,
    roles character varying(20)[]
);

ALTER TABLE ONLY users
    ADD CONSTRAINT users_pkey PRIMARY KEY (user_id);

ALTER TABLE ONLY profiles
    ADD CONSTRAINT profiles_user_id_foreign_key FOREIGN KEY (user_id) REFERENCES users(user_id);

