<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue';
import shaka from 'shaka-player';
import type { EpisodeDto, PlaybackSessionTicket, TrickplayCue } from '~/types/netflix';
import { useTrickplay } from '~/composables/useTrickplay';
import { useSessionStore } from '~/stores/session';

const props = defineProps<{ episode: EpisodeDto; ticket: PlaybackSessionTicket }>();
const emit = defineEmits<{ (e: 'nextEpisode'): void; (e: 'sessionRevoked', reason: string): void }>();

const sessionStore = useSessionStore();
const config = useRuntimeConfig();
const { loadTrickplayVtt, findCueAtMillisecond } = useTrickplay();

const videoEl = ref<HTMLVideoElement | null>(null);
const currentPosMs = ref<number>(props.ticket.resume_bookmark_ms);
const durationMs = ref<number>(props.episode.duration_ms);
const isPlaying = ref<boolean>(false);
const isAdBreakLocked = ref<boolean>(false);
const hoverCue = ref<TrickplayCue | null>(null);
const hoverLeftPx = ref<number>(0);
const audioTracks = ref<Array<{ id: number; language: string; label: string }>>([]);
const textTracks = ref<Array<{ id: number; language: string }>>([]);

let playerInstance: shaka.Player | null = null;
let heartbeatTimer: ReturnType<typeof setInterval> | null = null;
let lastReportedPosMs = props.ticket.resume_bookmark_ms;

const activeSkipMarker = computed(() => {
  return props.episode.markers.find(
    (m) => (m.marker_kind === 'INTRO' || m.marker_kind === 'RECAP') &&
      currentPosMs.value >= m.start_ms && currentPosMs.value <= m.end_ms
  ) ?? null;
});

const showNextEpisodePrompt = computed(() => {
  const credits = props.episode.markers.find((m) => m.marker_kind === 'CREDITS');
  if (credits) return currentPosMs.value >= credits.start_ms;
  return durationMs.value - currentPosMs.value <= 15000;
});

async function initializeShakaEngine(): Promise<void> {
  shaka.polyfill.installAll();
  if (!videoEl.value || !shaka.Player.isBrowserSupported()) return;

  playerInstance = new shaka.Player(videoEl.value);
  configureDrmEngine(playerInstance);

  const manifestUrl = props.ticket.signed_dash_manifest_url || props.ticket.signed_hls_manifest_url;
  await playerInstance.load(manifestUrl, props.ticket.resume_bookmark_ms / 1000);
  syncSelectableTracks();
  await loadTrickplayVtt(props.ticket.trickplay_vtt_url);
  startHeartbeatTelemetry();
}

function configureDrmEngine(player: shaka.Player): void {
  player.configure({
    drm: { servers: { 'com.widevine.alpha': config.public.drmLicenseUrl } },
    abr: { enabled: true, defaultBandwidthEstimate: 8000000 }
  });
  const netEngine = player.getNetworkingEngine();
  if (!netEngine) return;
  netEngine.registerRequestFilter((type, request) => {
    if (type === shaka.net.NetworkingEngine.RequestType.LICENSE) {
      request.headers['Authorization'] = `Bearer ${sessionStore.accessToken}`;
      request.headers['X-Drm-Session-Id'] = props.ticket.playback_session_id;
      request.headers['X-Drm-Token'] = props.ticket.drm_license_token;
    }
  });
}

function syncSelectableTracks(): void {
  if (!playerInstance) return;
  audioTracks.value = playerInstance.getVariantTracks().map((t) => ({
    id: t.id, language: t.language, label: t.label || t.language
  }));
  textTracks.value = playerInstance.getTextTracks().map((t) => ({ id: t.id, language: t.language }));
}

function startHeartbeatTelemetry(): void {
  heartbeatTimer = setInterval(async () => {
    if (!videoEl.value || !playerInstance) return;
    const stats = playerInstance.getStats();
    const deltaMs = Math.max(0, currentPosMs.value - lastReportedPosMs);
    lastReportedPosMs = currentPosMs.value;

    const res = await $fetch<{ keep_alive_allowed: boolean; revoke_reason: string }>(
      `${config.public.apiGatewayUrl}/api/v1/playback/heartbeat`,
      {
        method: 'POST',
        headers: sessionStore.authHeaders,
        body: {
          playback_session_id: props.ticket.playback_session_id,
          profile_id: sessionStore.activeProfile?.id ?? '',
          episode_id: props.episode.id,
          current_position_ms: currentPosMs.value,
          watched_delta_ms: deltaMs,
          current_bitrate_kbps: Math.round((stats.streamBandwidth ?? 0) / 1000),
          buffer_underrun_delta: 0,
          dropped_frames_delta: stats.droppedFrames ?? 0,
          is_paused: videoEl.value.paused
        }
      }
    );
    if (!res.keep_alive_allowed) {
      videoEl.value.pause();
      emit('sessionRevoked', res.revoke_reason);
    }
  }, 5000);
}

function handleTimeUpdate(): void {
  if (!videoEl.value) return;
  currentPosMs.value = Math.floor(videoEl.value.currentTime * 1000);
  const dueAd = props.ticket.ad_breaks.find(
    (ad) => Math.abs(currentPosMs.value - ad.cue_point_ms) < 600 && !isAdBreakLocked.value
  );
  if (dueAd) {
    isAdBreakLocked.value = true;
    setTimeout(() => { isAdBreakLocked.value = false; }, dueAd.duration_ms);
  }
}

function skipActiveMarker(): void {
  if (!videoEl.value || !activeSkipMarker.value || isAdBreakLocked.value) return;
  videoEl.value.currentTime = activeSkipMarker.value.end_ms / 1000;
}

function handleTimelineHover(e: MouseEvent): void {
  const bar = e.currentTarget as HTMLElement;
  const rect = bar.getBoundingClientRect();
  const ratio = Math.min(1, Math.max(0, (e.clientX - rect.left) / rect.width));
  const targetMs = Math.floor(ratio * durationMs.value);
  hoverLeftPx.value = e.clientX - rect.left;
  hoverCue.value = findCueAtMillisecond(targetMs);
}

function handleTimelineSeek(e: MouseEvent): void {
  if (!videoEl.value || isAdBreakLocked.value) return;
  const bar = e.currentTarget as HTMLElement;
  const rect = bar.getBoundingClientRect();
  const ratio = Math.min(1, Math.max(0, (e.clientX - rect.left) / rect.width));
  videoEl.value.currentTime = (ratio * durationMs.value) / 1000;
}

function switchAudioLanguage(lang: string): void { playerInstance?.selectAudioLanguage(lang); }
function switchSubtitleLanguage(lang: string): void {
  playerInstance?.selectTextLanguage(lang);
  playerInstance?.setTextTrackVisibility(true);
}

onMounted(() => { void initializeShakaEngine(); });
onBeforeUnmount(() => {
  if (heartbeatTimer) clearInterval(heartbeatTimer);
  void playerInstance?.destroy();
});
</script>

<template>
  <div class="relative w-full h-screen bg-black overflow-hidden select-none">
    <video ref="videoEl" class="w-full h-full object-contain" autoplay @timeupdate="handleTimeUpdate" @play="isPlaying = true" @pause="isPlaying = false" />

    <button
      v-if="activeSkipMarker && !isAdBreakLocked"
      class="absolute bottom-28 right-10 px-6 py-3 bg-white/90 text-black font-bold rounded shadow-lg hover:bg-white transition"
      @click="skipActiveMarker"
    >
      {{ activeSkipMarker.marker_kind === 'INTRO' ? '오프닝 건너뛰기' : '줄거리 건너뛰기' }}
    </button>

    <button
      v-if="showNextEpisodePrompt && !isAdBreakLocked"
      class="absolute bottom-28 right-10 px-6 py-3 bg-red-600 text-white font-bold rounded shadow-lg hover:bg-red-500 transition"
      @click="emit('nextEpisode')"
    >
      다음 화 재생 (5초 후 자동 이동)
    </button>

    <div class="absolute bottom-0 inset-x-0 p-6 bg-gradient-to-t from-black/90 to-transparent">
      <div
        v-if="hoverCue"
        class="absolute bottom-20 border-2 border-white rounded overflow-hidden shadow-xl pointer-events-none"
        :style="{
          left: `${Math.max(16, hoverLeftPx - hoverCue.w / 2)}px`,
          width: `${hoverCue.w}px`,
          height: `${hoverCue.h}px`,
          backgroundImage: `url(${hoverCue.spriteUrl})`,
          backgroundPosition: `-${hoverCue.x}px -${hoverCue.y}px`
        }"
      />

      <div
        class="relative w-full h-2 bg-zinc-700 rounded cursor-pointer mb-4"
        @mousemove="handleTimelineHover"
        @mouseleave="hoverCue = null"
        @click="handleTimelineSeek"
      >
        <div class="h-full bg-red-600 rounded" :style="{ width: `${(currentPosMs / Math.max(1, durationMs)) * 100}%` }" />
        <span
          v-for="ad in ticket.ad_breaks"
          :key="ad.cue_point_ms"
          class="absolute top-0 w-1.5 h-2 bg-yellow-400"
          :style="{ left: `${(ad.cue_point_ms / Math.max(1, durationMs)) * 100}%` }"
        />
      </div>

      <div class="flex items-center justify-between text-white text-sm">
        <div class="flex items-center gap-4">
          <span class="font-semibold">{{ episode.localized_title }}</span>
          <span v-if="isAdBreakLocked" class="px-2 py-0.5 bg-yellow-500 text-black font-bold rounded text-xs">
            광고 재생 중 (건너뛰기 불가)
          </span>
        </div>
        <div class="flex items-center gap-3">
          <select class="bg-zinc-900 border border-zinc-700 px-2 py-1 rounded" @change="switchAudioLanguage(($event.target as HTMLSelectElement).value)">
            <option v-for="a in audioTracks" :key="a.id" :value="a.language">오디오: {{ a.label }}</option>
          </select>
          <select class="bg-zinc-900 border border-zinc-700 px-2 py-1 rounded" @change="switchSubtitleLanguage(($event.target as HTMLSelectElement).value)">
            <option v-for="t in textTracks" :key="t.id" :value="t.language">자막: {{ t.language }}</option>
          </select>
        </div>
      </div>
    </div>
  </div>
</template>
