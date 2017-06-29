ALTER TABLE ONLY users
    DROP COLUMN IF EXISTS user_uuid;
ALTER TABLE ONLY booking_history
    DROP COLUMN IF EXISTS booking_history_uuid;
ALTER TABLE ONLY bookings
    DROP COLUMN IF EXISTS booking_uuid;
ALTER TABLE ONLY feedbacks
    DROP COLUMN IF EXISTS feedback_uuid;
ALTER TABLE ONLY messages
    DROP COLUMN IF EXISTS message_uuid;
ALTER TABLE ONLY notifications
    DROP COLUMN IF EXISTS notification_uuid;
ALTER TABLE ONLY payment_requests
    DROP COLUMN IF EXISTS payment_request_uuid;
ALTER TABLE ONLY profiles
    DROP COLUMN IF EXISTS profile_uuid;
ALTER TABLE ONLY rooms
    DROP COLUMN IF EXISTS room_uuid;
ALTER TABLE ONLY sda_types
    DROP COLUMN IF EXISTS sda_type_uuid;
ALTER TABLE ONLY servicecategory
    DROP COLUMN IF EXISTS servicecategory_uuid;
ALTER TABLE ONLY serviceinputtype
    DROP COLUMN IF EXISTS serviceinputtype_uuid;
ALTER TABLE ONLY tags
    DROP COLUMN IF EXISTS tag_uuid;