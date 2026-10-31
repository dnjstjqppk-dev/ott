import { defineEventHandler, readBody } from 'h3';

interface PlaybackRequestBody {
  episode_id?: string;
  device_id?: string;
  country_code?: string;
}

export default defineEventHandler(async (event) => {
  const body = await readBody<PlaybackRequestBody>(event);
  const episodeId = body?.episode_id || 'title-sq2';

  const expiresAt = new Date(Date.now() + 4 * 3600 * 1000).toISOString();
  const sessionId = `ses_${Date.now()}_${Math.random().toString(36).substring(2, 9)}`;

  // HLS Master playlist stream with trickplay preview
  const streamBase = 'https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4';

  return {
    playback_session_id: sessionId,
    drm_license_token: `drm_tok_${sessionId}`,
    signed_hls_manifest_url: streamBase,
    signed_dash_manifest_url: streamBase,
    trickplay_vtt_url: 'https://cdn.netflix-clone.internal/trickplay/preview.vtt',
    max_resolution: '4K Ultra HD',
    is_ad_break_required: false,
    session_expires_at: expiresAt,
    heartbeat_interval_sec: 10,
    resume_bookmark_ms: 0,
    allowed_watermark_text: 'USER 2026-KR-456'
  };
});
