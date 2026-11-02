import { defineEventHandler } from 'h3';

export default defineEventHandler((event) => {
  const jobId = event.context.params?.id || 'enc-retry';

  return {
    job_id: jobId,
    episode_id: 'ep-sq2-01',
    state: 'PROCESSING',
    total_chunks: 142,
    completed_chunks: 48,
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
      }
    ]
  };
});
