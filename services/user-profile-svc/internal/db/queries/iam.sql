-- name: GetUserByEmail :one
SELECT id, email, password_hash, role, status, country_code,
       failed_login_attempts, locked_until
FROM users
WHERE email = $1 LIMIT 1;

-- name: IncrementLoginFailure :exec
UPDATE users
SET failed_login_attempts = failed_login_attempts + 1,
    locked_until = CASE WHEN failed_login_attempts + 1 >= 5 THEN NOW() + INTERVAL '15 minutes' ELSE locked_until END,
    updated_at = NOW()
WHERE id = $1;

-- name: ResetLoginFailure :exec
UPDATE users
SET failed_login_attempts = 0, locked_until = NULL, updated_at = NOW()
WHERE id = $1;

-- name: UpsertUserDevice :one
INSERT INTO user_devices (user_id, device_fingerprint, device_name, platform, drm_security_level, last_ip, last_active_at, is_revoked)
VALUES ($1, $2, $3, $4, $5, $6, NOW(), FALSE)
ON CONFLICT (user_id, device_fingerprint)
DO UPDATE SET
    device_name = EXCLUDED.device_name,
    drm_security_level = EXCLUDED.drm_security_level,
    last_ip = EXCLUDED.last_ip,
    last_active_at = NOW(),
    is_revoked = FALSE
RETURNING id, user_id, device_fingerprint, platform, drm_security_level, is_revoked;

-- name: InsertRefreshToken :one
INSERT INTO refresh_tokens (user_id, device_id, token_sha256, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, device_id, token_sha256, expires_at;

-- name: GetActiveRefreshTokenByHash :one
SELECT rt.id, rt.user_id, rt.device_id, rt.expires_at,
       u.role, u.status, u.country_code, d.device_fingerprint, d.is_revoked
FROM refresh_tokens rt
JOIN users u ON u.id = rt.user_id
JOIN user_devices d ON d.id = rt.device_id
WHERE rt.token_sha256 = $1
  AND rt.revoked_at IS NULL
  AND rt.expires_at > NOW()
LIMIT 1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = NOW(), replaced_by_token_id = $2
WHERE id = $1 AND revoked_at IS NULL;

-- name: RevokeDeviceAndTokens :exec
UPDATE user_devices
SET is_revoked = TRUE, last_active_at = NOW()
WHERE id = $2 AND user_id = $1;

-- name: RevokeAllTokensForDevice :exec
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE user_id = $1 AND device_id = $2 AND revoked_at IS NULL;

-- name: ListProfilesByUser :many
SELECT id, user_id, display_name, avatar_key, is_kids, max_maturity,
       pin_bcrypt_hash, preferred_audio_lang, preferred_subtitle_lang,
       autoplay_next_episode, autoplay_previews, subtitle_font_scale,
       subtitle_bg_opacity, subtitle_edge_style
FROM profiles
WHERE user_id = $1
ORDER BY created_at ASC;

-- name: CountProfilesByUser :one
SELECT COUNT(*)::INT FROM profiles WHERE user_id = $1;

-- name: CreateProfile :one
INSERT INTO profiles (
    user_id, display_name, avatar_key, is_kids, max_maturity,
    pin_bcrypt_hash, preferred_audio_lang, preferred_subtitle_lang,
    autoplay_next_episode, autoplay_previews
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, user_id, display_name, avatar_key, is_kids, max_maturity,
          pin_bcrypt_hash, preferred_audio_lang, preferred_subtitle_lang,
          autoplay_next_episode, autoplay_previews, subtitle_font_scale,
          subtitle_bg_opacity, subtitle_edge_style;

-- name: UpdateProfile :one
UPDATE profiles
SET display_name = $3, avatar_key = $4, is_kids = $5, max_maturity = $6,
    pin_bcrypt_hash = $7, preferred_audio_lang = $8, preferred_subtitle_lang = $9,
    autoplay_next_episode = $10, autoplay_previews = $11, updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING id, user_id, display_name, avatar_key, is_kids, max_maturity,
          pin_bcrypt_hash, preferred_audio_lang, preferred_subtitle_lang,
          autoplay_next_episode, autoplay_previews, subtitle_font_scale,
          subtitle_bg_opacity, subtitle_edge_style;

-- name: GetProfileByID :one
SELECT id, user_id, display_name, is_kids, max_maturity, pin_bcrypt_hash
FROM profiles
WHERE id = $1 LIMIT 1;

