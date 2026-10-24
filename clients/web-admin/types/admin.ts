export interface LicensingWindowRecord {
  id: string;
  title_id: string;
  country_code: string;
  starts_at: string;
  ends_at: string;
  allow_download: boolean;
  allow_avod_tier: boolean;
}

export interface BanditArtworkMetric {
  artwork_id: string;
  storage_url: string;
  segment_cluster_id: number;
  impressions_count: number;
  play_starts_count: number;
  ctr_pct: number;
  alpha_param: number;
  beta_param: number;
}

export interface RepresentationSummary {
  id: string;
  codec: 'H264_AVC' | 'H265_HEVC' | 'AV1';
  width_px: number;
  height_px: number;
  bitrate_kbps: number;
  harmonic_mean_vmaf: number;
  min_chunk_vmaf: number;
}

export interface EncodingJobRecord {
  job_id: string;
  episode_id: string;
  state: string;
  total_chunks: number;
  completed_chunks: number;
  master_hls_key: string;
  master_dash_key: string;
  error_message: string;
  representations: RepresentationSummary[];
}

export interface OpenConnectEdgeNode {
  id: string;
  node_code: string;
  region_code: string;
  status: 'HEALTHY' | 'DEGRADED' | 'DRAINING' | 'OFFLINE';
  current_bandwidth_gbps: number;
  cache_hit_ratio: number;
  active_streams: number;
}

export interface ImmutableAuditLogRecord {
  id: string;
  actor_user_id: string;
  actor_role: string;
  actor_ip: string;
  action_code: string;
  target_entity_type: string;
  target_entity_id: string;
  before_state: Record<string, unknown>;
  after_state: Record<string, unknown>;
  record_sha256: string;
  created_at: string;
}

