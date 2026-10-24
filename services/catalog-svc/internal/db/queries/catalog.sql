-- name: GetTitleByID :one
SELECT t.id, t.slug, t.type,
       COALESCE(t.localized_titles->>sqlc.arg(lang)::text, t.original_title)::text AS display_title,
       COALESCE(t.synopsis->>sqlc.arg(lang)::text, '')::text AS display_synopsis,
       t.release_year, t.maturity_rating, t.maturity_reasons,
       t.is_netflix_original, t.max_resolution, t.supports_hdr10,
       t.supports_dolby_vision, t.supports_dolby_atmos,
       COALESCE(pti.in_my_list, FALSE) AS in_my_list,
       COALESCE(pti.rating::text, 'NONE')::text AS user_rating
FROM titles t
LEFT JOIN profile_title_interactions pti
  ON pti.title_id = t.id AND pti.profile_id = sqlc.arg(profile_id)::uuid
WHERE t.id = sqlc.arg(title_id)::uuid AND t.state = 'PUBLISHED'
LIMIT 1;

-- name: GetActiveArtworksForTitle :many
SELECT kind, storage_key
FROM artworks
WHERE title_id = $1 AND language_code = $2 AND is_active = TRUE;

-- name: ListSeasonsByTitle :many
SELECT s.id, s.season_number,
       COALESCE(s.localized_names->>sqlc.arg(column_2)::text, 'Season ' || s.season_number::text)::text AS display_name,
       COUNT(e.id)::INT AS episode_count
FROM seasons s
LEFT JOIN episodes e ON e.season_id = s.id AND e.state = 'PUBLISHED'
WHERE s.title_id = sqlc.arg(title_id)::uuid
GROUP BY s.id
ORDER BY s.season_number ASC;

-- name: ListEpisodesBySeason :many
SELECT e.id, e.title_id, e.season_id, e.episode_number,
       COALESCE(e.localized_titles->>sqlc.arg(lang)::text, 'Episode ' || e.episode_number::text)::text AS display_title,
       COALESCE(e.synopsis->>sqlc.arg(lang)::text, '')::text AS display_synopsis,
       e.duration_ms, e.still_image_key,
       COALESCE(wb.bookmark_ms, 0)::BIGINT AS user_bookmark_ms
FROM episodes e
LEFT JOIN watch_bookmarks wb
  ON wb.episode_id = e.id AND wb.profile_id = sqlc.arg(profile_id)::uuid
WHERE e.season_id = sqlc.arg(season_id)::uuid AND e.state = 'PUBLISHED'
ORDER BY e.episode_number ASC;

-- name: ListTimecodeMarkersByEpisode :many
SELECT marker_kind, start_ms, end_ms, start_frame, end_frame
FROM timecode_markers
WHERE episode_id = $1
ORDER BY start_ms ASC;

-- name: ListCreditsByTitle :many
SELECT p.id AS person_id, p.full_name, tc.credit_role, COALESCE(tc.character_name, '') AS character_name
FROM title_credits tc
JOIN people p ON p.id = tc.person_id
WHERE tc.title_id = $1
ORDER BY tc.billing_order ASC
LIMIT 20;

-- name: ListSimilarTitles :many
SELECT t.id, t.slug, t.type,
       COALESCE(t.localized_titles->>$2, t.original_title)::text AS display_title,
       t.release_year, t.maturity_rating, t.max_resolution
FROM titles t
JOIN title_genres tg ON tg.title_id = t.id
WHERE tg.genre_id IN (SELECT tg2.genre_id FROM title_genres tg2 WHERE tg2.title_id = $1)
  AND t.id != $1
  AND t.state = 'PUBLISHED'
GROUP BY t.id
ORDER BY t.popularity_score DESC
LIMIT 12;

-- name: CheckActiveLicenseWindow :one
SELECT allow_download, allow_avod_tier, ends_at
FROM licensing_windows
WHERE title_id = $1
  AND country_code = $2
  AND starts_at <= NOW()
  AND ends_at > NOW()
ORDER BY ends_at DESC
LIMIT 1;

-- name: UpsertTitleRecord :one
INSERT INTO titles (
    partner_org_id, slug, type, original_title, localized_titles,
    synopsis, release_year, maturity_rating, maturity_reasons,
    is_netflix_original, state
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'PUBLISHED')
ON CONFLICT (slug) DO UPDATE SET
    original_title = EXCLUDED.original_title,
    localized_titles = EXCLUDED.localized_titles,
    synopsis = EXCLUDED.synopsis,
    maturity_rating = EXCLUDED.maturity_rating,
    maturity_reasons = EXCLUDED.maturity_reasons,
    updated_at = NOW()
RETURNING id, slug, type, original_title, release_year, maturity_rating, is_netflix_original, max_resolution;
