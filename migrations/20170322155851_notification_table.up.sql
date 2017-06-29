CREATE SEQUENCE     notifications_notification_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE notifications(
	notification_id bigint DEFAULT nextval('notifications_notification_id_seq'::regclass) NOT NULL,
	notification_type text,
	sender_id bigint,
	receiver_id bigint,
	entity_id bigint,
	notification_string text
);

ALTER TABLE ONLY notifications
	ADD CONSTRAINT notifications_pkey PRIMARY KEY (notification_id);

ALTER TABLE ONLY notifications
	ADD CONSTRAINT notifications_sender_id_foreign_key FOREIGN KEY (sender_id) REFERENCES users(user_id);

ALTER TABLE ONLY notifications
	ADD CONSTRAINT notifications_receiver_id_foreign_key FOREIGN KEY (receiver_id) REFERENCES users(user_id);