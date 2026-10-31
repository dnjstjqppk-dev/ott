<!-- clients/web-user/components/NetflixPlayerModal.vue -->
<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue';
import type { TitleDto, PlaybackSessionTicket } from '~/types/netflix';

const props = defineProps<{
  title: TitleDto | null;
  isOpen: boolean;
  ticket?: PlaybackSessionTicket | null;
}>();
const emit = defineEmits<{ (e: 'close'): void }>();

const videoRef = ref<HTMLVideoElement | null>(null);
const isPlaying = ref<boolean>(true);
const isMuted = ref<boolean>(false);
const currentTime = ref<number>(0);
const duration = ref<number>(100);
const isControlsVisible = ref<boolean>(true);
let hideTimer: ReturnType<typeof setTimeout> | null = null;

function resetControlsTimeout(): void {
  isControlsVisible.value = true;
  if (hideTimer) clearTimeout(hideTimer);
  hideTimer = setTimeout(() => {
    if (isPlaying.value) isControlsVisible.value = false;
  }, 3500);
}

function togglePlay(): void {
  if (!videoRef.value) return;
  if (videoRef.value.paused) {
    void videoRef.value.play();
    isPlaying.value = true;
  } else {
    videoRef.value.pause();
    isPlaying.value = false;
  }
}

function seekRelative(seconds: number): void {
  if (!videoRef.value) return;
  videoRef.value.currentTime = Math.max(
    0,
    Math.min(duration.value, videoRef.value.currentTime + seconds)
  );
}

function handleTimeUpdate(): void {
  if (!videoRef.value) return;
  currentTime.value = videoRef.value.currentTime;
  duration.value = videoRef.value.duration || 100;
}

function handleScrub(e: Event): void {
  const target = e.target as HTMLInputElement;
  if (videoRef.value) {
    videoRef.value.currentTime = parseFloat(target.value);
  }
}

function toggleMute(): void {
  if (!videoRef.value) return;
  isMuted.value = !isMuted.value;
  videoRef.value.muted = isMuted.value;
}

function toggleFullscreen(): void {
  if (!document.fullscreenElement) {
    void document.documentElement.requestFullscreen();
  } else {
    void document.exitFullscreen();
  }
}

function formatTime(seconds: number): string {
  const m = Math.floor(seconds / 60);
  const s = Math.floor(seconds % 60);
  return `${m}:${s < 10 ? '0' : ''}${s}`;
}

onMounted(() => {
  resetControlsTimeout();
});

onBeforeUnmount(() => {
  if (hideTimer) clearTimeout(hideTimer);
});
</script>

<template>
  <div
    v-if="isOpen && title"
    class="fixed inset-0 z-50 bg-black flex flex-col justify-between overflow-hidden cursor-pointer"
    @mousemove="resetControlsTimeout"
  >
    <video
      ref="videoRef"
      autoplay
      playsinline
      :src="
        props.ticket?.signed_hls_manifest_url ||
        'https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4'
      "
      class="absolute inset-0 w-full h-full object-cover z-0"
      @timeupdate="handleTimeUpdate"
      @click="togglePlay"
    />

    <div
      class="relative z-10 p-6 flex items-center justify-between bg-gradient-to-b from-black/80 to-transparent transition-opacity duration-300"
      :class="isControlsVisible ? 'opacity-100' : 'opacity-0 pointer-events-none'"
    >
      <div class="flex items-center gap-4">
        <button
          class="p-2 hover:bg-white/20 rounded-full transition text-white"
          aria-label="뒤로가기"
          @click="$emit('close')"
        >
          <svg class="w-6 h-6 fill-none stroke-current stroke-2" viewBox="0 0 24 24">
            <path d="M19 12H5M12 19l-7-7 7-7" />
          </svg>
        </button>
        <div>
          <h2 class="text-lg md:text-xl font-bold text-white tracking-tight">
            {{ title.localized_title }}
          </h2>
          <span class="text-xs text-zinc-400">제1화 · 4K Ultra HD</span>
        </div>
      </div>
      <div class="px-3 py-1 bg-red-600/80 text-white rounded text-xs font-bold">
        NETFLIX PLAYBACK
      </div>
    </div>

    <div
      class="relative z-10 p-6 pb-8 bg-gradient-to-t from-black/90 via-black/40 to-transparent transition-opacity duration-300 space-y-3"
      :class="isControlsVisible ? 'opacity-100' : 'opacity-0 pointer-events-none'"
    >
      <div class="flex items-center gap-3">
        <input
          type="range"
          min="0"
          :max="duration"
          step="0.1"
          :value="currentTime"
          class="w-full accent-red-600 h-1.5 bg-zinc-700 rounded-lg cursor-pointer"
          @input="handleScrub"
        />
        <span class="text-xs text-zinc-300 font-mono w-24 text-right">
          {{ formatTime(currentTime) }} / {{ formatTime(duration) }}
        </span>
      </div>

      <div class="flex items-center justify-between text-white">
        <div class="flex items-center gap-4">
          <button class="p-2 hover:text-red-500 transition text-lg" @click="togglePlay">
            {{ isPlaying ? '❚❚' : '▶' }}
          </button>
          <button
            class="p-2 text-xs font-bold hover:text-zinc-300"
            @click="seekRelative(-10)"
          >
            ↺ 10초
          </button>
          <button
            class="p-2 text-xs font-bold hover:text-zinc-300"
            @click="seekRelative(10)"
          >
            ↻ 10초
          </button>
          <button class="p-2 hover:text-zinc-300 transition" @click="toggleMute">
            {{ isMuted ? '🔇' : '🔊' }}
          </button>
        </div>

        <div class="flex items-center gap-4 text-xs font-medium">
          <span class="px-2 py-0.5 border border-zinc-600 rounded text-zinc-300">
            한국어 [음성 및 자막]
          </span>
          <button class="hover:text-zinc-300" @click="toggleFullscreen">
            ⛶ 전체화면
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
