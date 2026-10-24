CREATE TYPE title_type AS ENUM ('MOVIE', 'SERIES');
CREATE TYPE release_state AS ENUM ('DRAFT', 'INGESTING', 'QC_PENDING', 'READY', 'PUBLISHED', 'ARCHIVED');
CREATE TYPE artwork_type AS ENUM ('BOXART_VERTICAL', 'BOXART_HORIZONTAL', 'BILLBOARD_HERO', 'TITLE_LOGO_PNG', 'TOP10_CARD');
CREATE TYPE marker_type AS ENUM ('RECAP', 'INTRO', 'CREDITS', 'POST_CREDITS', 'AD_BREAK_CUE');

CREATE TABLE titles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    partner_org_id UUID NOT NULL REFERENCES partner_organizations(id),
    slug VARCHAR(255) NOT NULL UNIQUE,
    type title_type NOT NULL,
    original_title VARCHAR(255) NOT NULL,
    localized_titles JSONB NOT NULL DEFAULT '{}',
    synopsis JSONB NOT NULL DEFAULT '{}',
    release_year SMALLINT NOT NULL,
    maturity_rating maturity_level NOT NULL,
    maturity_reasons TEXT[] NOT NULL DEFAULT '{}',
    is_netflix_original BOOLEAN NOT NULL DEFAULT FALSE,
    original_language VARCHAR(10) NOT NULL DEFAULT 'ko',
    max_resolution VARCHAR(16) NOT NULL DEFAULT '4K_UHD',
    supports_hdr10 BOOLEAN NOT NULL DEFAULT FALSE,
    supports_dolby_vision BOOLEAN NOT NULL DEFAULT FALSE,
    supports_dolby_atmos BOOLEAN NOT NULL DEFAULT FALSE,
    state release_state NOT NULL DEFAULT 'DRAFT',
    popularity_score DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_titles_state_pop ON titles(state, popularity_score DESC);
CREATE INDEX idx_titles_type ON titles(type);

CREATE TABLE seasons (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title_id UUID NOT NULL REFERENCES titles(id) ON DELETE CASCADE,
    season_number SMALLINT NOT NULL,
    localized_names JSONB NOT NULL DEFAULT '{}',
    synopsis JSONB NOT NULL DEFAULT '{}',
    release_date DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (title_id, season_number)
);

CREATE TABLE episodes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title_id UUID NOT NULL REFERENCES titles(id) ON DELETE CASCADE,
    season_id UUID REFERENCES seasons(id) ON DELETE CASCADE,
    episode_number SMALLINT NOT NULL DEFAULT 1,
    localized_titles JSONB NOT NULL DEFAULT '{}',
    synopsis JSONB NOT NULL DEFAULT '{}',
    duration_ms BIGINT NOT NULL CHECK (duration_ms > 0),
    frame_rate_num INT NOT NULL DEFAULT 24000,
    frame_rate_den INT NOT NULL DEFAULT 1001,
    still_image_key TEXT NOT NULL,
    state release_state NOT NULL DEFAULT 'DRAFT',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (season_id, episode_number)
);

CREATE INDEX idx_episodes_title ON episodes(title_id);

CREATE TABLE timecode_markers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    episode_id UUID NOT NULL REFERENCES episodes(id) ON DELETE CASCADE,
    marker_kind marker_type NOT NULL,
    start_ms BIGINT NOT NULL CHECK (start_ms >= 0),
    end_ms BIGINT NOT NULL CHECK (end_ms > start_ms),
    start_frame BIGINT NOT NULL CHECK (start_frame >= 0),
    end_frame BIGINT NOT NULL CHECK (end_frame > start_frame),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (episode_id, marker_kind, start_ms)
);

CREATE INDEX idx_timecode_markers_episode ON timecode_markers(episode_id);

CREATE TABLE genres (
    id SMALLSERIAL PRIMARY KEY,
    code VARCHAR(64) NOT NULL UNIQUE,
    localized_names JSONB NOT NULL
);

CREATE TABLE title_genres (
    title_id UUID NOT NULL REFERENCES titles(id) ON DELETE CASCADE,
    genre_id SMALLINT NOT NULL REFERENCES genres(id) ON DELETE CASCADE,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (title_id, genre_id)
);

CREATE TABLE people (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    full_name VARCHAR(255) NOT NULL,
    localized_names JSONB NOT NULL DEFAULT '{}',
    birth_date DATE,
    headshot_key TEXT
);

CREATE TABLE title_credits (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title_id UUID NOT NULL REFERENCES titles(id) ON DELETE CASCADE,
    person_id UUID NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    credit_role VARCHAR(32) NOT NULL CHECK (credit_role IN ('ACTOR', 'DIRECTOR', 'WRITER', 'CREATOR', 'PRODUCER')),
    character_name VARCHAR(255),
    billing_order SMALLINT NOT NULL DEFAULT 1
);

CREATE INDEX idx_title_credits_title ON title_credits(title_id, billing_order);
CREATE INDEX idx_title_credits_person ON title_credits(person_id);

CREATE TABLE artworks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title_id UUID NOT NULL REFERENCES titles(id) ON DELETE CASCADE,
    kind artwork_type NOT NULL,
    language_code VARCHAR(10) NOT NULL DEFAULT 'ko',
    focal_person_id UUID REFERENCES people(id),
    theme_tags TEXT[] NOT NULL DEFAULT '{}',
    storage_key TEXT NOT NULL,
    width_px INT NOT NULL,
    height_px INT NOT NULL,
    sha256_hash CHAR(64) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_artworks_title_kind ON artworks(title_id, kind, language_code) WHERE is_active = TRUE;

CREATE TABLE licensing_windows (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title_id UUID NOT NULL REFERENCES titles(id) ON DELETE CASCADE,
    country_code CHAR(2) NOT NULL,
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL CHECK (ends_at > starts_at),
    allow_download BOOLEAN NOT NULL DEFAULT TRUE,
    allow_avod_tier BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (title_id, country_code, starts_at)
);

CREATE INDEX idx_licensing_active ON licensing_windows(country_code, starts_at, ends_at);
