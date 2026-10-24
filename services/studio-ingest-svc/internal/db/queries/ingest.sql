-- name: CreateIngestSession :one
INSERT INTO ingest_sessions (
    episode_id, uploaded_by, partner_org_id, file_name,
    source_container, total_byte_size, chunk_size_bytes,
    expected_sha256, minio_bucket, minio_object_key, status, expires_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'INITIATED', NOW() + INTERVAL '7 days')
RETURNING id, episode_id, total_byte_size, uploaded_byte_offset, chunk_size_bytes,
          expected_sha256, minio_bucket, minio_object_key, status;

-- name: GetIngestSessionByID :one
SELECT id, episode_id, uploaded_by, partner_org_id, file_name,
       source_container, total_byte_size, uploaded_byte_offset,
       chunk_size_bytes, expected_sha256, computed_sha256,
       minio_bucket, minio_object_key, status
FROM ingest_sessions
WHERE id = $1 LIMIT 1;

-- name: InsertIngestChunk :exec
INSERT INTO ingest_chunks (
    ingest_session_id, chunk_index, start_byte, end_byte,
    byte_length, chunk_sha256, etag
) VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: AdvanceIngestByteOffset :one
UPDATE ingest_sessions
SET uploaded_byte_offset = $2,
    status = 'UPLOADING'
WHERE id = $1 AND uploaded_byte_offset = $3
RETURNING uploaded_byte_offset;

-- name: MarkIngestCompleted :exec
UPDATE ingest_sessions
SET computed_sha256 = $2,
    status = 'COMPLETED',
    completed_at = NOW()
WHERE id = $1;

-- name: ListIngestChunksOrdered :many
SELECT chunk_index, start_byte, end_byte, byte_length, chunk_sha256, etag
FROM ingest_chunks
WHERE ingest_session_id = $1
ORDER BY chunk_index ASC;

-- name: CreateInitialQcReport :one
INSERT INTO qc_reports (
    ingest_session_id, episode_id, verdict, detected_width, detected_height,
    detected_fps_num, detected_fps_den, color_space, hdr_format,
    audio_channels, integrated_lufs, true_peak_dbtp
) VALUES ($1, $2, 'RUNNING', 3840, 2160, 24000, 1001, 'BT2020', 'HDR10', '7.1.4_ATMOS', -24.00, -2.10)
RETURNING id;

-- name: GetLatestQcReportByEpisode :one
SELECT id, ingest_session_id, episode_id, verdict, detected_width,
       detected_height, detected_fps_num, detected_fps_den,
       color_space, hdr_format, audio_channels, integrated_lufs, true_peak_dbtp
FROM qc_reports
WHERE episode_id = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: ListQcIssuesByReport :many
SELECT id, issue_type, severity, start_ms, end_ms, start_frame, end_frame,
       measured_value, threshold_value, COALESCE(reviewer_comment, '') AS reviewer_comment, is_resolved
FROM qc_issues
WHERE qc_report_id = $1
ORDER BY start_ms ASC;

-- name: UpdateQcReportDecision :one
UPDATE qc_reports
SET verdict = $2, reviewed_by = $3, review_notes = $4, completed_at = NOW()
WHERE id = $1
RETURNING id, ingest_session_id, episode_id, verdict, detected_width,
          detected_height, detected_fps_num, detected_fps_den,
          color_space, hdr_format, audio_channels, integrated_lufs, true_peak_dbtp;

-- name: DeleteTimedTextCuesByTrack :exec
DELETE FROM timed_text_cues WHERE track_id = $1;

-- name: InsertTimedTextCue :exec
INSERT INTO timed_text_cues (
    track_id, cue_sequence, start_ms, end_ms, text_lines,
    computed_cps, max_line_length, violates_shot_change, violates_cps_limit
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: UpdateTimedTextTrackStorage :exec
UPDATE timed_text_tracks
SET author_user_id = $2, webvtt_storage_key = $3, verdict = $4, updated_at = NOW()
WHERE id = $1;

