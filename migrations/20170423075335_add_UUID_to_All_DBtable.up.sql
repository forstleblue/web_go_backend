CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

ALTER TABLE ONLY users
    ADD COLUMN user_uuid UUID DEFAULT uuid_generate_v1();
ALTER TABLE ONLY profiles
    ADD COLUMN profile_uuid UUID DEFAULT uuid_generate_v1();
ALTER TABLE ONLY bookings
    ADD COLUMN booking_uuid UUID DEFAULT uuid_generate_v1();
ALTER TABLE ONLY booking_history
    ADD COLUMN booking_history_uuid UUID DEFAULT uuid_generate_v1();
ALTER TABLE ONLY notifications
    ADD COLUMN notification_uuid UUID DEFAULT uuid_generate_v1();
ALTER TABLE ONLY messages
    ADD COLUMN message_uuid UUID DEFAULT uuid_generate_v1();
ALTER TABLE ONLY feedbacks
    ADD COLUMN feedback_uuid UUID DEFAULT uuid_generate_v1();
ALTER TABLE ONLY payment_requests
    ADD COLUMN payment_request_uuid UUID DEFAULT uuid_generate_v1();
ALTER TABLE ONLY rooms
    ADD COLUMN room_uuid UUID DEFAULT uuid_generate_v1();
ALTER TABLE ONLY sda_types
    ADD COLUMN sda_type_uuid UUID DEFAULT uuid_generate_v1();
ALTER TABLE ONLY servicecategory
    ADD COLUMN servicecategory_uuid UUID DEFAULT uuid_generate_v1();
ALTER TABLE ONLY serviceinputtype
    ADD COLUMN serviceinputtype_uuid UUID DEFAULT uuid_generate_v1();
ALTER TABLE ONLY tags
    ADD COLUMN tag_uuid UUID DEFAULT uuid_generate_v1();