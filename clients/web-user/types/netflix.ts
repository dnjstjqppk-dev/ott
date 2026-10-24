export interface ProfileDto {
  id: string;
  user_id: string;
  display_name: string;
  avatar_key: string;
  is_kids: boolean;
  max_maturity: 'ALL' | 'AGE_7' | 'AGE_12' | 'AGE_15' | 'AGE_19';
  has_pin: boolean;
  preferred_audio_lang: string;
  preferred_subtitle_lang: string;
  autoplay_next_episode: boolean;
  autoplay_previews: boolean;
  subtitle_font_scale: number;
  subtitle_bg_opacity: number;
  subtitle_edge_style: string;
}

export interface TimecodeMarkerDto {
  marker_kind: 'RECAP' | 'INTRO' | 'CREDITS' | 'POST_CREDITS' | 'AD_BREAK_CUE';
  start_ms: number;
  end_ms: number;
  start_frame: number;
  end_frame: number;
}

export interface EpisodeDto {
  id: string;
  title_id: string;
  season_id: string;
  episode_number: number;
  localized_title: string;
  localized_synopsis: string;
  duration_ms: number;
  still_image_url: string;
  user_bookmark_ms: number;
  markers: TimecodeMarkerDto[];
}

export interface TitleDto {
  id: string;
  slug: string;
  type: 'MOVIE' | 'SERIES';
  localized_title: string;
  localized_synopsis: string;
  release_year: number;
  maturity_rating: string;
  maturity_reasons: string[];
  is_netflix_original: boolean;
  max_resolution: string;
  supports_hdr10: boolean;
  supports_dolby_vision: boolean;
  supports_dolby_atmos: boolean;
  hero_artwork_url: string;
  logo_png_url: string;
  boxart_url: string;
  match_score_pct: number;
  in_my_list: boolean;
  user_rating: string;
}

export interface LoLoMoItemDto {
  title_id: string;
  personalized_artwork_url: string;
  match_score_pct: number;
  rank_number: number;
  bookmark_ms: number;
  duration_ms: number;
}

export interface LoLoMoRowDto {
  row_id: string;
  row_index: number;
  row_type: string;
  header_text: string;
  items: LoLoMoItemDto[];
}

export interface AggregatedHomeResponse {
  lolomo: {
    snapshot_id: string;
    hero_billboard: LoLoMoItemDto;
    rows: LoLoMoRowDto[];
  };
  subscription: {
    subscription_id: string;
    plan_code: 'STANDARD_WITH_ADS' | 'STANDARD' | 'PREMIUM_UHD';
    max_concurrent_streams: number;
    max_resolution: string;
    has_ads: boolean;
  };
}

export interface AdBreakInsertionDto {
  cue_point_ms: number;
  duration_ms: number;
  ad_hls_url: string;
  ad_dash_url: string;
}

export interface PlaybackSessionTicket {
  playback_session_id: string;
  signed_hls_manifest_url: string;
  signed_dash_manifest_url: string;
  trickplay_vtt_url: string;
  drm_license_token: string;
  widevine_pssh_base64: string;
  resume_bookmark_ms: number;
  ad_breaks: AdBreakInsertionDto[];
}

export interface TrickplayCue {
  startMs: number;
  endMs: number;
  spriteUrl: string;
  x: number;
  y: number;
  w: number;
  h: number;
}

