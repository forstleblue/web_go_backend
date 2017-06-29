DROP INDEX IF EXISTS idx_gin_api_configuration;

ALTER TABLE ONLY platforms
    DROP COLUMN IF EXISTS api_configuration;