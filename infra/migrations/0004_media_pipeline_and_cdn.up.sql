CREATE TYPE encoding_job_state AS ENUM ('QUEUED', 'CHUNKING', 'COMPLEXITY_ANALYSIS', 'ENCODING', 'VMAF_SCORING', 'PACKAGING', 'COMPLETED', 'FAILED_DLQ');
CREATE TYPE video_codec AS ENUM ('H264_AVC', 'H265_HEVC', 'AV1');

CREATE TABLE drm_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    episode_id UUID NOT NULL UNIQUE REFERENCES episodes(id) ON DELETE CASCADE,
    key_id_hex CHAR(32) NOT NULL UNIQUE,
    content_key_encrypted BYTEA NOT NULL,
    iv_hex CHAR(32) NOT NULL,
    cenc_scheme VARCHAR(8) NOT NULL DEFAULT 'cenc' CHECK (cenc_scheme IN ('cenc', 'cbcs')),
    widevine_pssh_base64 TEXT NOT NULL,
    fairplay_skd_uri TEXT NOT NULL,
    rotation_version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE encoding_jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    episode_id UUID NOT NULL REFERENCES episodes(id) ON DELETE CASCADE,
    ingest_session_id UUID NOT NULL REFERENCES ingest_sessions(id),
    worker_node_id VARCHAR(128),
    state encoding_job_state NOT NULL DEFAULT 'QUEUED',
    total_chunks INT NOT NULL DEFAULT 0,
    completed_chunks INT NOT NULL DEFAULT 0,
    min_target_vmaf NUMERIC(5,2) NOT NULL DEFAULT 92.00,
    master_hls_key TEXT,
    master_dash_key TEXT,
    error_message TEXT,
    retry_count SMALLINT NOT NULL DEFAULT 0,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_encoding_jobs_state ON encoding_jobs(state, created_at);

CREATE TABLE video_scene_chunks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    encoding_job_id UUID NOT NULL REFERENCES encoding_jobs(id) ON DELETE CASCADE,
    episode_id UUID NOT NULL REFERENCES episodes(id) ON DELETE CASCADE,
    chunk_index INT NOT NULL CHECK (chunk_index >= 0),
    start_ms BIGINT NOT NULL CHECK (start_ms >= 0),
    end_ms BIGINT NOT NULL CHECK (end_ms > start_ms),
    start_frame BIGINT NOT NULL CHECK (start_frame >= 0),
    end_frame BIGINT NOT NULL CHECK (end_frame > start_frame),
    spatial_complexity NUMERIC(6,3) NOT NULL,
    temporal_complexity NUMERIC(6,3) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (encoding_job_id, chunk_index)
);

CREATE TABLE encoded_representations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    encoding_job_id UUID NOT NULL REFERENCES encoding_jobs(id) ON DELETE CASCADE,
    episode_id UUID NOT NULL REFERENCES episodes(id) ON DELETE CASCADE,
    codec video_codec NOT NULL,
    profile_name VARCHAR(32) NOT NULL,
    width_px INT NOT NULL,
    height_px INT NOT NULL,
    bitrate_kbps INT NOT NULL CHECK (bitrate_kbps > 0),
    Peak_bitrate_kbps INT NOT NULL CHECK (Peak_bitrate_kbps >= bitrate_kbps),
    harmonic_mean_vmaf NUMERIC(5,2) NOT NULL,
    min_chunk_vmaf NUMERIC(5,2) NOT NULL,
    init_segment_key TEXT NOT NULL,
    init_segment_byte_size BIGINT NOT NULL,
    init_segment_sha256 CHAR(64) NOT NULL,
    playlist_key TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (encoding_job_id, codec, height_px, bitrate_kbps)
);

CREATE INDEX idx_encoded_rep_episode ON encoded_representations(episode_id, codec);

CREATE TABLE cmaf_segments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    representation_id UUID NOT NULL REFERENCES encoded_representations(id) ON DELETE CASCADE,
    scene_chunk_id UUID NOT NULL REFERENCES video_scene_chunks(id) ON DELETE CASCADE,
    segment_index INT NOT NULL CHECK (segment_index >= 0),
    duration_ms INT NOT NULL CHECK (duration_ms > 0),
    storage_key TEXT NOT NULL,
    byte_size BIGINT NOT NULL CHECK (byte_size > 0),
    byte_offset BIGINT NOT NULL DEFAULT 0 CHECK (byte_offset >= 0),
    sha256_hash CHAR(64) NOT NULL,
    chunk_vmaf_score NUMERIC(5,2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (representation_id, segment_index)
);

CREATE INDEX idx_cmaf_segments_lookup ON cmaf_segments(representation_id, segment_index);

CREATE TABLE trickplay_tracks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    episode_id UUID NOT NULL UNIQUE REFERENCES episodes(id) ON DELETE CASCADE,
    interval_ms INT NOT NULL DEFAULT 2000,
    thumb_width_px INT NOT NULL DEFAULT 320,
    thumb_height_px INT NOT NULL DEFAULT 180,
    columns_per_sheet SMALLINT NOT NULL DEFAULT 10,
    rows_per_sheet SMALLINT NOT NULL DEFAULT 10,
    vtt_manifest_key TEXT NOT NULL,
    sprite_sheet_keys TEXT[] NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE open_connect_edge_nodes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    node_code VARCHAR(64) NOT NULL UNIQUE,
    isp_asn INT NOT NULL,
    region_code VARCHAR(16) NOT NULL,
    country_code CHAR(2) NOT NULL,
    base_url TEXT NOT NULL,
    max_bandwidth_gbps NUMERIC(6,2) NOT NULL,
    current_bandwidth_gbps NUMERIC(6,2) NOT NULL DEFAULT 0.0,
    cache_hit_ratio NUMERIC(5,4) NOT NULL DEFAULT 0.0,
    active_streams INT NOT NULL DEFAULT 0,
    status VARCHAR(24) NOT NULL DEFAULT 'HEALTHY' CHECK (status IN ('HEALTHY', 'DEGRADED', 'DRAINING', 'OFFLINE')),
    last_heartbeat_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_edge_nodes_routing ON open_connect_edge_nodes(country_code, status, current_bandwidth_gbps);
