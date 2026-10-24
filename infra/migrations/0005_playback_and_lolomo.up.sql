CREATE TYPE interaction_rating AS ENUM ('NONE', 'THUMBS_DOWN', 'THUMBS_UP', 'DOUBLE_THUMBS_UP');

CREATE TABLE playback_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    episode_id UUID NOT NULL REFERENCES episodes(id) ON DELETE CASCADE,
    device_id UUID NOT NULL REFERENCES user_devices(id) ON DELETE CASCADE,
    assigned_edge_node_id UUID NOT NULL REFERENCES open_connect_edge_nodes(id),
    manifest_hmac_signature CHAR(64) NOT NULL,
    drm_license_token TEXT NOT NULL UNIQUE,
    current_position_ms BIGINT NOT NULL DEFAULT 0,
    current_bitrate_kbps INT NOT NULL DEFAULT 0,
    buffer_underrun_count INT NOT NULL DEFAULT 0,
    dropped_frames_count INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_heartbeat_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMPTZ
);

CREATE INDEX idx_playback_active_user ON playback_sessions(user_id, is_active) WHERE is_active = TRUE;
CREATE INDEX idx_playback_heartbeat ON playback_sessions(last_heartbeat_at) WHERE is_active = TRUE;

CREATE TABLE watch_bookmarks (
    profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    title_id UUID NOT NULL REFERENCES titles(id) ON DELETE CASCADE,
    episode_id UUID NOT NULL REFERENCES episodes(id) ON DELETE CASCADE,
    bookmark_ms BIGINT NOT NULL CHECK (bookmark_ms >= 0),
    duration_ms BIGINT NOT NULL CHECK (duration_ms > 0),
    is_completed BOOLEAN NOT NULL DEFAULT FALSE,
    is_hidden_from_row BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (profile_id, episode_id)
);

CREATE INDEX idx_watch_bookmarks_continue ON watch_bookmarks(profile_id, updated_at DESC)
    WHERE is_completed = FALSE AND is_hidden_from_row = FALSE;

CREATE TABLE watch_history_ledger (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    title_id UUID NOT NULL REFERENCES titles(id) ON DELETE CASCADE,
    episode_id UUID NOT NULL REFERENCES episodes(id) ON DELETE CASCADE,
    watched_ms_delta BIGINT NOT NULL CHECK (watched_ms_delta > 0),
    country_code CHAR(2) NOT NULL,
    recorded_date DATE NOT NULL DEFAULT CURRENT_DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_watch_ledger_title_date ON watch_history_ledger(title_id, recorded_date);

CREATE TABLE profile_title_interactions (
    profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    title_id UUID NOT NULL REFERENCES titles(id) ON DELETE CASCADE,
    in_my_list BOOLEAN NOT NULL DEFAULT FALSE,
    my_list_added_at TIMESTAMPTZ,
    rating interaction_rating NOT NULL DEFAULT 'NONE',
    rated_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (profile_id, title_id)
);

CREATE INDEX idx_interactions_mylist ON profile_title_interactions(profile_id, my_list_added_at DESC) WHERE in_my_list = TRUE;

CREATE TABLE lolomo_snapshots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    context_hash CHAR(64) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_lolomo_profile ON lolomo_snapshots(profile_id, expires_at DESC);

CREATE TABLE lolomo_rows (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    snapshot_id UUID NOT NULL REFERENCES lolomo_snapshots(id) ON DELETE CASCADE,
    row_index SMALLINT NOT NULL,
    row_type VARCHAR(64) NOT NULL CHECK (row_type IN (
        'CONTINUE_WATCHING', 'TOP_10_COUNTRY', 'MY_LIST', 'BECAUSE_YOU_WATCHED',
        'GENRE_COLLECTION', 'NETFLIX_ORIGINALS', 'TRENDING_NOW', 'NEW_RELEASES'
    )),
    localized_header JSONB NOT NULL,
    UNIQUE (snapshot_id, row_index)
);

CREATE TABLE lolomo_row_items (
    row_id UUID NOT NULL REFERENCES lolomo_rows(id) ON DELETE CASCADE,
    item_index SMALLINT NOT NULL,
    title_id UUID NOT NULL REFERENCES titles(id) ON DELETE CASCADE,
    personalized_artwork_id UUID NOT NULL REFERENCES artworks(id),
    match_score_pct SMALLINT NOT NULL CHECK (match_score_pct BETWEEN 50 AND 99),
    rank_number SMALLINT,
    PRIMARY KEY (row_id, item_index)
);

CREATE TABLE artwork_bandit_stats (
    artwork_id UUID NOT NULL REFERENCES artworks(id) ON DELETE CASCADE,
    segment_cluster_id SMALLINT NOT NULL,
    impressions_count BIGINT NOT NULL DEFAULT 0,
    play_starts_count BIGINT NOT NULL DEFAULT 0,
    qualified_plays_count BIGINT NOT NULL DEFAULT 0,
    alpha_param DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    beta_param DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (artwork_id, segment_cluster_id)
);
