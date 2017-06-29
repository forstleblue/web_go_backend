 CREATE SEQUENCE     rooms_room_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE rooms(
	room_id bigint DEFAULT nextval('rooms_room_id_seq'::regclass) NOT NULL,
	room_owner bigint,
	users bigint[],
	created timestamp without time zone
);

ALTER TABLE ONLY rooms
	ADD CONSTRAINT rooms_pkey PRIMARY KEY (room_id);

ALTER TABLE ONLY rooms
    ADD CONSTRAINT rooms_room_owner_foreign_key FOREIGN KEY (room_owner) REFERENCES users(user_id);
