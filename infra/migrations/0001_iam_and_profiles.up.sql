CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TYPE account_role AS ENUM ('USER', 'PARTNER_ADMIN', 'PARTNER_ENGINEER', 'PARTNER_LOCALIZER', 'SUPER_ADMIN', 'SRE_OPERATOR');
CREATE TYPE account_status AS ENUM ('ONBOARDING', 'ACTIVE', 'PAST_DUE', 'SUSPENDED', 'CANCELED');
CREATE TYPE maturity_level AS ENUM ('ALL', 'AGE_7', 'AGE_12', 'AGE_15', 'AGE_19');
CREATE TYPE device_platform AS ENUM ('WEB', 'IOS', 'ANDROID', 'SMART_TV_WEBOS', 'SMART_TV_TIZEN', 'ANDROID_TV', 'DESKTOP');

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    phone_e164 VARCHAR(32) UNIQUE,
    role account_role NOT NULL DEFAULT 'USER',
    status account_status NOT NULL DEFAULT 'ONBOARDING',
    country_code CHAR(2) NOT NULL DEFAULT 'KR',
    mfa_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    mfa_secret TEXT,
    failed_login_attempts INT NOT NULL DEFAULT 0,
    locked_until TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);

CREATE TABLE profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    display_name VARCHAR(64) NOT NULL,
    avatar_key VARCHAR(128) NOT NULL,
    is_kids BOOLEAN NOT NULL DEFAULT FALSE,
    max_maturity maturity_level NOT NULL DEFAULT 'AGE_19',
    pin_bcrypt_hash TEXT,
    preferred_audio_lang VARCHAR(10) NOT NULL DEFAULT 'ko',
    preferred_subtitle_lang VARCHAR(10) NOT NULL DEFAULT 'ko',
    autoplay_next_episode BOOLEAN NOT NULL DEFAULT TRUE,
    autoplay_previews BOOLEAN NOT NULL DEFAULT TRUE,
    subtitle_font_scale NUMERIC(3,2) NOT NULL DEFAULT 1.00,
    subtitle_bg_opacity NUMERIC(3,2) NOT NULL DEFAULT 0.75,
    subtitle_edge_style VARCHAR(32) NOT NULL DEFAULT 'DROP_SHADOW',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_profiles_user_id ON profiles(user_id);

CREATE TABLE profile_title_restrictions (
    profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    title_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (profile_id, title_id)
);

CREATE TABLE user_devices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_fingerprint VARCHAR(128) NOT NULL,
    device_name VARCHAR(128) NOT NULL,
    platform device_platform NOT NULL,
    drm_security_level VARCHAR(16) NOT NULL DEFAULT 'L1',
    last_ip INET NOT NULL,
    last_active_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_revoked BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, device_fingerprint)
);

CREATE INDEX idx_user_devices_user_active ON user_devices(user_id, is_revoked);

CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id UUID NOT NULL REFERENCES user_devices(id) ON DELETE CASCADE,
    token_sha256 CHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    replaced_by_token_id UUID REFERENCES refresh_tokens(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_lookup ON refresh_tokens(token_sha256) WHERE revoked_at IS NULL;

CREATE TABLE tv_activation_codes (
    code CHAR(8) PRIMARY KEY,
    device_fingerprint VARCHAR(128) NOT NULL,
    platform device_platform NOT NULL,
    ip_address INET NOT NULL,
    bound_user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE partner_organizations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    legal_name VARCHAR(255) NOT NULL UNIQUE,
    organization_type VARCHAR(64) NOT NULL CHECK (organization_type IN ('STUDIO', 'POST_HOUSE', 'LOCALIZATION_VENDOR', 'AD_AGENCY')),
    country_code CHAR(2) NOT NULL,
    tax_registration_no VARCHAR(64) NOT NULL,
    watermark_template TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE partner_memberships (
    partner_org_id UUID NOT NULL REFERENCES partner_organizations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    allowed_languages TEXT[] NOT NULL DEFAULT '{}',
    can_upload_raw BOOLEAN NOT NULL DEFAULT FALSE,
    can_approve_qc BOOLEAN NOT NULL DEFAULT FALSE,
    can_view_settlements BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (partner_org_id, user_id)
);
