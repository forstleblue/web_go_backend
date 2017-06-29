ALTER TABLE ONLY platforms
    ADD COLUMN api_configuration jsonb;

CREATE INDEX idx_gin_api_configuration ON platforms USING gin (api_configuration);