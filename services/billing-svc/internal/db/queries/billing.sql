-- name: GetUserSubscriptionWithPlan :one
SELECT s.id AS subscription_id, s.plan_code, s.current_period_end,
       p.max_concurrent_streams, p.max_download_devices,
       p.max_resolution, p.has_ads
FROM subscriptions s
JOIN subscription_plans p ON p.code = s.plan_code
WHERE s.user_id = $1
LIMIT 1;

-- name: UpsertUserSubscriptionPlan :one
INSERT INTO subscriptions (user_id, plan_code, current_period_start, current_period_end)
VALUES ($1, $2, NOW(), NOW() + INTERVAL '1 month')
ON CONFLICT (user_id) DO UPDATE SET
    plan_code = EXCLUDED.plan_code,
    updated_at = NOW()
RETURNING id;

-- name: ListPartnerContractsForOrg :many
SELECT id, partner_org_id, title_id, model, minimum_guarantee_krw,
       rate_per_qualified_hour_krw, revenue_share_bps
FROM partner_contracts
WHERE partner_org_id = $1
  AND effective_from <= CURRENT_DATE
  AND effective_to >= CURRENT_DATE;

-- name: SumWatchedMillisecondsForTitleMonth :one
SELECT COALESCE(SUM(watched_ms_delta), 0)::BIGINT AS total_watched_ms
FROM watch_history_ledger
WHERE title_id = sqlc.arg(title_id)::uuid
  AND TO_CHAR(recorded_date, 'YYYY-MM') = sqlc.arg(column_2)::text;

-- name: UpsertPartnerSettlementLedger :one
INSERT INTO partner_settlement_ledgers (
    contract_id, partner_org_id, title_id, settlement_month,
    total_watched_ms, qualified_hours_viewed, completion_rate_pct,
    gross_payout_krw, withholding_tax_krw, net_payout_krw, status
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'CALCULATED')
ON CONFLICT (contract_id, settlement_month) DO UPDATE SET
    total_watched_ms = EXCLUDED.total_watched_ms,
    qualified_hours_viewed = EXCLUDED.qualified_hours_viewed,
    gross_payout_krw = EXCLUDED.gross_payout_krw,
    withholding_tax_krw = EXCLUDED.withholding_tax_krw,
    net_payout_krw = EXCLUDED.net_payout_krw
RETURNING id, title_id, qualified_hours_viewed, gross_payout_krw, net_payout_krw, status;

-- name: InsertImmutableAuditLog :exec
INSERT INTO immutable_audit_logs (
    actor_user_id, actor_role, actor_ip, action_code,
    target_entity_type, target_entity_id, before_state, after_state, record_sha256
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

