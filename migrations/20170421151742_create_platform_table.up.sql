CREATE TABLE platform (
    platform_id uuid NOT NULL,
    name text NOT NULL,
    profile_type text NOT NULL,
    widget_access boolean DEFAULT false,
    created timestamp without time zone,
    updated timestamp without time zone
);


ALTER TABLE ONLY platform
    ADD CONSTRAINT   platform_pkey   PRIMARY KEY (platform_id);