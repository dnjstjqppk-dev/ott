import { defineEventHandler } from 'h3';

export default defineEventHandler(() => {
  return {
    windows: [
      {
        id: 'win-sq2-kr',
        title_id: '00000000-0000-0000-0000-000000000001',
        country_code: 'KR',
        starts_at: '2026-10-01T00:00:00Z',
        ends_at: '2028-10-01T00:00:00Z',
        allow_download: true,
        allow_avod_tier: false
      },
      {
        id: 'win-glory-global',
        title_id: '00000000-0000-0000-0000-000000000002',
        country_code: 'GLOBAL',
        starts_at: '2023-01-01T00:00:00Z',
        ends_at: '2030-01-01T00:00:00Z',
        allow_download: true,
        allow_avod_tier: true
      },
      {
        id: 'win-parasite-kr',
        title_id: '00000000-0000-0000-0000-000000000003',
        country_code: 'KR',
        starts_at: '2024-01-01T00:00:00Z',
        ends_at: '2027-12-31T23:59:59Z',
        allow_download: true,
        allow_avod_tier: true
      }
    ],
    bandit_metrics: [
      {
        artwork_id: 'art-sq2-neon-hero',
        storage_url: '/posters/hero-squidgame2.jpg',
        segment_cluster_id: 101,
        impressions_count: 1425890,
        play_starts_count: 249530,
        ctr_pct: 17.5,
        alpha_param: 2496,
        beta_param: 11763
      },
      {
        artwork_id: 'art-glory-dramatic',
        storage_url: '/posters/the-glory.jpg',
        segment_cluster_id: 102,
        impressions_count: 984210,
        play_starts_count: 157470,
        ctr_pct: 16.0,
        alpha_param: 1575,
        beta_param: 8267
      },
      {
        artwork_id: 'art-parasite-modern',
        storage_url: '/posters/parasite.jpg',
        segment_cluster_id: 103,
        impressions_count: 612400,
        play_starts_count: 88180,
        ctr_pct: 14.4,
        alpha_param: 882,
        beta_param: 5242
      }
    ],
    encoding_jobs: [
      {
        job_id: 'enc-sq2-ep01-4k',
        episode_id: 'ep-sq2-01',
        state: 'COMPLETED',
        total_chunks: 142,
        completed_chunks: 142,
        master_hls_key: 'hls/sq2/ep01/master.m3u8',
        master_dash_key: 'dash/sq2/ep01/manifest.mpd',
        error_message: '',
        representations: [
          {
            id: 'rep-av1-4k-uhd',
            codec: 'AV1',
            width_px: 3840,
            height_px: 2160,
            bitrate_kbps: 8400,
            harmonic_mean_vmaf: 97.4,
            min_chunk_vmaf: 94.2
          },
          {
            id: 'rep-hevc-1080p-hdr',
            codec: 'H265_HEVC',
            width_px: 1920,
            height_px: 1080,
            bitrate_kbps: 4200,
            harmonic_mean_vmaf: 96.1,
            min_chunk_vmaf: 92.8
          },
          {
            id: 'rep-h264-720p-sd',
            codec: 'H264_AVC',
            width_px: 1280,
            height_px: 720,
            bitrate_kbps: 1800,
            harmonic_mean_vmaf: 92.5,
            min_chunk_vmaf: 89.1
          }
        ]
      }
    ],
    edge_nodes: [
      {
        id: 'oca-icn-01',
        node_code: 'KR-ICN-IX-01',
        region_code: 'ap-northeast-2',
        status: 'HEALTHY',
        current_bandwidth_gbps: 84.6,
        cache_hit_ratio: 98.7,
        active_streams: 32410
      },
      {
        id: 'oca-nrt-02',
        node_code: 'JP-NRT-IX-02',
        region_code: 'ap-northeast-1',
        status: 'HEALTHY',
        current_bandwidth_gbps: 62.1,
        cache_hit_ratio: 97.9,
        active_streams: 21850
      },
      {
        id: 'oca-sin-01',
        node_code: 'SG-SIN-IX-01',
        region_code: 'ap-southeast-1',
        status: 'HEALTHY',
        current_bandwidth_gbps: 48.3,
        cache_hit_ratio: 96.8,
        active_streams: 15400
      }
    ],
    audit_logs: [
      {
        id: 'aud-001',
        actor_user_id: 'adm-netflix-sre-001',
        actor_role: 'SUPER_ADMIN',
        actor_ip: '10.240.0.12',
        action_code: 'CATALOG_LICENSE_WINDOW_EXTEND',
        target_entity_type: 'CATALOG_WINDOW',
        target_entity_id: 'win-sq2-kr',
        before_state: { country: 'KR', valid_until: '2027-10-01' },
        after_state: { country: 'KR', valid_until: '2028-10-01' },
        record_sha256: 'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855',
        created_at: new Date(Date.now() - 3600000).toISOString()
      },
      {
        id: 'aud-002',
        actor_user_id: 'adm-netflix-sre-001',
        actor_role: 'SUPER_ADMIN',
        actor_ip: '10.240.0.12',
        action_code: 'CDN_EDGE_OPTIMIZE',
        target_entity_type: 'EDGE_APPLIANCE',
        target_entity_id: 'oca-icn-01',
        before_state: { status: 'DRAINING' },
        after_state: { status: 'HEALTHY' },
        record_sha256: 'a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e',
        created_at: new Date(Date.now() - 7200000).toISOString()
      }
    ]
  };
});
