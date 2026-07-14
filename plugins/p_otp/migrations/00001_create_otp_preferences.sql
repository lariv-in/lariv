-- +goose Up
CREATE TABLE IF NOT EXISTS otp_preferences (
    id                             BIGSERIAL PRIMARY KEY,
    created_at                     TIMESTAMPTZ,
    updated_at                     TIMESTAMPTZ,
    deleted_at                     TIMESTAMPTZ,
    sms_otp_template_id            TEXT,
    otp_template_id                TEXT,
    msg91_auth_key                 TEXT,
    sms_otp_field_name             TEXT,
    sms_otp_extra_fields           TEXT,
    email_otp_template_string      TEXT,
    smtp_host                      TEXT,
    smtp_port                      TEXT,
    smtp_username                  TEXT,
    smtp_password                  TEXT,
    smtp_from                      TEXT
);

CREATE INDEX IF NOT EXISTS idx_otp_preferences_deleted_at ON otp_preferences (deleted_at);

-- +goose Down
DROP TABLE IF EXISTS otp_preferences;
