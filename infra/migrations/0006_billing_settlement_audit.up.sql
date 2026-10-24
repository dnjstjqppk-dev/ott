CREATE TYPE plan_tier AS ENUM ('STANDARD_WITH_ADS', 'STANDARD', 'PREMIUM_UHD');
CREATE TYPE contract_model AS ENUM ('ORIGINAL_BUYOUT', 'MG_PLUS_REVENUE_SHARE', 'HOURS_VIEWED_POOL');

CREATE TABLE subscription_plans (
    code plan_tier PRIMARY KEY,
    display_name JSONB NOT NULL,
    monthly_price_krw INT NOT NULL CHECK (monthly_price_krw >= 0),
    monthly_price_usd_cents INT NOT NULL CHECK (monthly_price_usd_cents >= 0),
    max_concurrent_streams SMALLINT NOT NULL,
    max_download_devices SMALLINT NOT NULL,
    max_resolution VARCHAR(16) NOT NULL,
    supports_spatial_audio BOOLEAN NOT NULL DEFAULT FALSE,
    has_ads BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    plan_code plan_tier NOT NULL REFERENCES subscription_plans(code),
    pending_plan_code plan_tier REFERENCES subscription_plans(code),
    current_period_start TIMESTAMPTZ NOT NULL,
    current_period_end TIMESTAMPTZ NOT NULL,
    cancel_at_period_end BOOLEAN NOT NULL DEFAULT FALSE,
    auto_renew BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE payment_methods (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    pg_provider VARCHAR(32) NOT NULL,
    billing_key_encrypted TEXT NOT NULL,
    card_brand VARCHAR(32) NOT NULL,
    card_last4 CHAR(4) NOT NULL,
    expiry_month SMALLINT NOT NULL,
    expiry_year SMALLINT NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE billing_invoices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    subscription_id UUID NOT NULL REFERENCES subscriptions(id),
    payment_method_id UUID REFERENCES payment_methods(id),
    amount_krw INT NOT NULL,
    tax_amount_krw INT NOT NULL,
    currency CHAR(3) NOT NULL DEFAULT 'KRW',
    pg_transaction_id VARCHAR(128) UNIQUE,
    status VARCHAR(24) NOT NULL CHECK (status IN ('PAID', 'FAILED', 'REFUNDED')),
    paid_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE avod_ad_creatives (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    advertiser_name VARCHAR(255) NOT NULL,
    duration_ms INT NOT NULL CHECK (duration_ms > 0),
    target_maturity maturity_level NOT NULL DEFAULT 'AGE_15',
    target_genres SMALLINT[] NOT NULL DEFAULT '{}',
    hls_manifest_key TEXT NOT NULL,
    dash_manifest_key TEXT NOT NULL,
    cpm_rate_krw INT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE partner_contracts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    partner_org_id UUID NOT NULL REFERENCES partner_organizations(id),
    title_id UUID NOT NULL REFERENCES titles(id),
    model contract_model NOT NULL,
    minimum_guarantee_krw BIGINT NOT NULL DEFAULT 0,
    rate_per_qualified_hour_krw INT NOT NULL DEFAULT 0,
    revenue_share_bps INT NOT NULL DEFAULT 0 CHECK (revenue_share_bps BETWEEN 0 AND 10000),
    effective_from DATE NOT NULL,
    effective_to DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (partner_org_id, title_id)
);

CREATE TABLE partner_settlement_ledgers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    contract_id UUID NOT NULL REFERENCES partner_contracts(id),
    partner_org_id UUID NOT NULL REFERENCES partner_organizations(id),
    title_id UUID NOT NULL REFERENCES titles(id),
    settlement_month CHAR(7) NOT NULL,
    total_watched_ms BIGINT NOT NULL DEFAULT 0,
    qualified_hours_viewed NUMERIC(12,2) NOT NULL DEFAULT 0.00,
    completion_rate_pct NUMERIC(5,2) NOT NULL DEFAULT 0.00,
    gross_payout_krw BIGINT NOT NULL DEFAULT 0,
    withholding_tax_krw BIGINT NOT NULL DEFAULT 0,
    net_payout_krw BIGINT NOT NULL DEFAULT 0,
    statement_pdf_key TEXT,
    status VARCHAR(24) NOT NULL DEFAULT 'CALCULATED' CHECK (status IN ('CALCULATED', 'APPROVED', 'PAID')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (contract_id, settlement_month)
);

CREATE TABLE immutable_audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    actor_user_id UUID NOT NULL REFERENCES users(id),
    actor_role account_role NOT NULL,
    actor_ip INET NOT NULL,
    action_code VARCHAR(128) NOT NULL,
    target_entity_type VARCHAR(64) NOT NULL,
    target_entity_id UUID NOT NULL,
    before_state JSONB,
    after_state JSONB,
    record_sha256 CHAR(64) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_entity ON immutable_audit_logs(target_entity_type, target_entity_id, created_at DESC);
