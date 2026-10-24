CREATE TYPE ingest_status AS ENUM ('INITIATED', 'UPLOADING', 'HASH_VERIFYING', 'COMPLETED', 'FAILED', 'ABORTED');
CREATE TYPE qc_verdict AS ENUM ('PENDING', 'RUNNING', 'PASSED', 'WARNING', 'REJECTED', 'OVERRIDDEN_APPROVED');
CREATE TYPE qc_issue_severity AS ENUM ('INFO', 'WARNING', 'CRITICAL_BLOCKER');

CREATE TABLE ingest_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    episode_id UUID NOT NULL REFERENCES episodes(id) ON DELETE CASCADE,
    uploaded_by UUID NOT NULL REFERENCES users(id),
    partner_org_id UUID NOT NULL REFERENCES partner_organizations(id),
    file_name VARCHAR(512) NOT NULL,
    source_container VARCHAR(32) NOT NULL CHECK (source_container IN ('MOV_PRORES', 'MXF_IMF', 'MKV', 'MP4', 'WAV_ADM')),
    total_byte_size BIGINT NOT NULL CHECK (total_byte_size > 0),
    uploaded_byte_offset BIGINT NOT NULL DEFAULT 0 CHECK (uploaded_byte_offset >= 0),
    chunk_size_bytes INT NOT NULL DEFAULT 16777216,
    expected_sha256 CHAR(64) NOT NULL,
    computed_sha256 CHAR(64),
    minio_bucket VARCHAR(128) NOT NULL DEFAULT 'nf-raw-ingest',
    minio_object_key TEXT NOT NULL,
    status ingest_status NOT NULL DEFAULT 'INITIATED',
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX idx_ingest_episode ON ingest_sessions(episode_id, status);

CREATE TABLE ingest_chunks (
    ingest_session_id UUID NOT NULL REFERENCES ingest_sessions(id) ON DELETE CASCADE,
    chunk_index INT NOT NULL CHECK (chunk_index >= 0),
    start_byte BIGINT NOT NULL CHECK (start_byte >= 0),
    end_byte BIGINT NOT NULL CHECK (end_byte >= start_byte),
    byte_length INT NOT NULL CHECK (byte_length > 0),
    chunk_sha256 CHAR(64) NOT NULL,
    etag VARCHAR(128) NOT NULL,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (ingest_session_id, chunk_index)
);

CREATE TABLE imf_packages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ingest_session_id UUID NOT NULL UNIQUE REFERENCES ingest_sessions(id) ON DELETE CASCADE,
    cpl_uuid UUID NOT NULL,
    cpl_xml_key TEXT NOT NULL,
    pkl_xml_key TEXT NOT NULL,
    assetmap_xml_key TEXT NOT NULL,
    video_mxf_keys TEXT[] NOT NULL,
    audio_mxf_keys TEXT[] NOT NULL,
    parsed_metadata JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE qc_reports (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ingest_session_id UUID NOT NULL REFERENCES ingest_sessions(id) ON DELETE CASCADE,
    episode_id UUID NOT NULL REFERENCES episodes(id) ON DELETE CASCADE,
    verdict qc_verdict NOT NULL DEFAULT 'PENDING',
    detected_width INT NOT NULL,
    detected_height INT NOT NULL,
    detected_fps_num INT NOT NULL,
    detected_fps_den INT NOT NULL,
    color_space VARCHAR(32) NOT NULL,
    hdr_format VARCHAR(32) NOT NULL DEFAULT 'SDR_REC709',
    audio_channels VARCHAR(32) NOT NULL,
    integrated_lufs NUMERIC(5,2) NOT NULL,
    true_peak_dbtp NUMERIC(5,2) NOT NULL,
    max_CLL INT,
    max_FALL INT,
    reviewed_by UUID REFERENCES users(id),
    review_notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX idx_qc_reports_episode ON qc_reports(episode_id, verdict);

CREATE TABLE qc_issues (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    qc_report_id UUID NOT NULL REFERENCES qc_reports(id) ON DELETE CASCADE,
    issue_type VARCHAR(64) NOT NULL CHECK (issue_type IN (
        'BLACK_FRAME', 'FREEZE_FRAME', 'AUDIO_SILENCE', 'AUDIO_CLIPPING',
        'LUFS_OUT_OF_RANGE', 'SUBTITLE_SYNC_DRIFT', 'CPS_EXCEEDED', 'SHOT_CHANGE_VIOLATION', 'DEAD_PIXEL'
    )),
    severity qc_issue_severity NOT NULL,
    start_ms BIGINT NOT NULL CHECK (start_ms >= 0),
    end_ms BIGINT NOT NULL CHECK (end_ms >= start_ms),
    start_frame BIGINT NOT NULL CHECK (start_frame >= 0),
    end_frame BIGINT NOT NULL CHECK (end_frame >= start_frame),
    measured_value VARCHAR(128) NOT NULL,
    threshold_value VARCHAR(128) NOT NULL,
    reviewer_comment TEXT,
    is_resolved BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_qc_issues_report ON qc_issues(qc_report_id, severity);

CREATE TABLE timed_text_tracks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    episode_id UUID NOT NULL REFERENCES episodes(id) ON DELETE CASCADE,
    language_code VARCHAR(10) NOT NULL,
    track_type VARCHAR(32) NOT NULL CHECK (track_type IN ('SUBTITLE', 'SDH_CC', 'FORCED_NARRATIVE')),
    Assigned_partner_org_id UUID REFERENCES partner_organizations(id),
    author_user_id UUID REFERENCES users(id),
    verdict qc_verdict NOT NULL DEFAULT 'PENDING',
    webvtt_storage_key TEXT,
    ttml_storage_key TEXT,
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (episode_id, language_code, track_type, version)
);

CREATE TABLE timed_text_cues (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    track_id UUID NOT NULL REFERENCES timed_text_tracks(id) ON DELETE CASCADE,
    cue_sequence INT NOT NULL,
    start_ms BIGINT NOT NULL CHECK (start_ms >= 0),
    end_ms BIGINT NOT NULL CHECK (end_ms > start_ms),
    text_lines TEXT[] NOT NULL,
    computed_cps NUMERIC(5,2) NOT NULL,
    max_line_length SMALLINT NOT NULL,
    violates_shot_change BOOLEAN NOT NULL DEFAULT FALSE,
    violates_cps_limit BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (track_id, cue_sequence)
);

CREATE INDEX idx_timed_text_cues_track ON timed_text_cues(track_id, start_ms);
