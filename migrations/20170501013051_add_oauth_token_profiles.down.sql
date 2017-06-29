ALTER TABLE ONLY profiles
    DROP COLUMN IF EXISTS oauth_token;

ALTER TABLE ONLY profiles
    DROP COLUMN IF EXISTS oauth_expiry;