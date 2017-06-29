CREATE TABLE widgets(
    widget_id  uuid NOT NULL,
    profile_id bigint,
    type text,
    configuration json,
    created timestamp without time zone,
    updated timestamp without time zone   
);


ALTER TABLE ONLY widgets
    ADD CONSTRAINT   widget_pkey   PRIMARY KEY (widget_id);

ALTER TABLE ONLY widgets
    ADD CONSTRAINT   widget_profile_id_foreign_key FOREIGN KEY (profile_id) REFERENCES profiles(profile_id);