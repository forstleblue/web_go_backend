CREATE SEQUENCE     sda_types_sda_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
CREATE TABLE    sda_types (
    sda_id      bigint DEFAULT nextval('sda_types_sda_id_seq'::regclass) NOT NULL,
    ref_id      bigint,
    profile_type char(1),
    sda_list    text[]
);

ALTER TABLE ONLY sda_types
    ADD CONSTRAINT sda_types_pkey PRIMARY KEY (sda_id);

ALTER TABLE ONLY sda_types
    ADD CONSTRAINT sda_type_ref_id_foreign_key FOREIGN KEY (ref_id)  REFERENCES servicecategory(service_id);
