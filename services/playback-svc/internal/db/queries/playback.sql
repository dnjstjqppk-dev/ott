-- name: CountActivePlaybackSessionsByUser :one
SELECT COUNT(*)::INT FROM playback_sessions
WHERE user_id = $1 AND is_active = TRUE;

-- name: GetActiveSubscriptionStreamLimit :one
SELECT p.max_concurrent_streams
FROM subscriptions s
JOIN subscription_plans p ON p.code = s.plan_code
WHERE s.user_id = $1
LIMIT 1;

-- name: GetDrmKeyByEpisode :one
SELECT id, key_id_hex, content_key_encrypted, iv_hex, cenc_scheme, widevine_pssh_base64, fairplay_skd_uri
FROM drm_keys
WHERE episode_id = $1
LIMIT 1;

-- name: CreatePlaybackSession :one
INSERT INTO playback_sessions (
    user_id, profile_id, episode_id, device_id, assigned_edge_node_id,
    manifest_hmac_signature, drm_license_token, current_position_ms, is_active
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, TRUE)
RETURNING id, assigned_edge_node_id, drm_license_token;

-- name: UpdatePlaybackHeartbeat :exec
UPDATE playback_sessions
SET current_position_ms = $2,
    current_bitrate_kbps = $3,
    buffer_underrun_count = buffer_underrun_count + $4,
    dropped_frames_count = dropped_frames_count + $5,
    last_heartbeat_at = NOW()
WHERE id = $1 AND is_active = TRUE;

-- name: AppendWatchHistoryDelta :exec
INSERT INTO watch_history_ledger (profile_id, title_id, episode_id, watched_ms_delta, country_code, recorded_date)
VALUES ($1, $2, $3, $4, $5, CURRENT_DATE);

-- name: UpsertWatchBookmark :exec
INSERT INTO watch_bookmarks (profile_id, title_id, episode_id, bookmark_ms, duration_ms, is_completed, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW())
ON CONFLICT (profile_id, episode_id) DO UPDATE SET
    bookmark_ms = EXCLUDED.bookmark_ms,
    is_completed = EXCLUDED.is_completed,
    updated_at = NOW();

-- name: TerminatePlaybackSession :exec
UPDATE playback_sessions
SET is_active = FALSE, current_position_ms = $2, ended_at = NOW()
WHERE id = $1 AND is_active = TRUE;

-- name: GetEpisodeTitleAndDuration :one
SELECT title_id, duration_ms FROM episodes WHERE id = $1 LIMIT 1;

