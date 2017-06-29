CREATE SEQUENCE messages_message_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE messages (
    message_id bigint DEFAULT nextval('messages_message_id_seq'::regclass) NOT NULL,
    room_id bigint,
    message_text text,
    user_id bigint,
    created timestamp without time zone,
    unread bigint[]
);

ALTER TABLE ONLY messages
    ADD CONSTRAINT messages_pkey PRIMARY KEY (message_id);

ALTER TABLE ONLY messages
    ADD CONSTRAINT messages_room_id_foreign_key FOREIGN KEY (room_id) REFERENCES rooms(room_id);

ALTER TABLE ONLY messages
    ADD CONSTRAINT messages_user_id_foreign_key FOREIGN KEY (user_id) REFERENCES users(user_id);
