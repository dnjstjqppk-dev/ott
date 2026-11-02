<!-- clients/web-user/components/NetflixTitleDetailModal.vue -->
<script setup lang="ts">
import type { TitleDto } from '~/types/netflix';

const props = defineProps<{
  title: TitleDto | null;
  isOpen: boolean;
  isInMyList: boolean;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'play', titleId: string): void;
  (e: 'toggleMyList', titleId: string): void;
}>();

function onImgError(e: Event): void {
  const el = e.target as HTMLImageElement;
  if (el) el.src = '/posters/fallback.jpg';
}
</script>

<template>
  <div
    v-if="isOpen && title"
    class="fixed inset-0 z-50 bg-black/80 backdrop-blur-md flex items-center justify-center p-4 md:p-8 overflow-y-auto animate-fade-in"
    @click.self="emit('close')"
  >
    <div
      class="relative w-full max-w-4xl bg-zinc-900 rounded-xl overflow-hidden shadow-2xl border border-zinc-800 text-white max-h-[90vh] flex flex-col"
    >
      <button
        class="absolute top-4 right-4 z-20 w-9 h-9 bg-zinc-900/80 hover:bg-zinc-800 rounded-full flex items-center justify-center text-white text-lg transition"
        aria-label="닫기"
        @click="emit('close')"
      >
        ✕
      </button>

      <div class="relative w-full h-64 md:h-96 flex-none bg-zinc-950">
        <img
          :src="title.hero_artwork_url || '/posters/hero-squidgame2.jpg'"
          :alt="title.localized_title"
          class="w-full h-full object-cover"
          onerror="this.onerror=null;this.src='/posters/fallback.jpg';"
          @error="onImgError"
        />
        <div class="absolute inset-0 bg-gradient-to-t from-zinc-900 via-transparent to-transparent" />
        <div class="absolute inset-0 bg-gradient-to-r from-zinc-900/90 via-transparent to-transparent" />

        <div class="absolute bottom-6 left-6 md:left-10 space-y-3 z-10 max-w-lg">
          <h2 class="text-2xl md:text-4xl font-black tracking-tight drop-shadow-md">
            {{ title.localized_title }}
          </h2>
          <div class="flex items-center gap-3 pt-1">
            <button
              class="px-6 py-2.5 bg-white text-black font-bold text-sm rounded flex items-center gap-2 hover:bg-zinc-200 transition shadow"
              @click="emit('play', title.id)"
            >
              ▶ 재생
            </button>
            <button
              class="px-4 py-2.5 bg-zinc-800/80 hover:bg-zinc-700 text-white font-medium text-sm rounded flex items-center gap-2 border border-zinc-600 transition"
              @click="emit('toggleMyList', title.id)"
            >
              <span>{{ isInMyList ? '✓ 찜한 콘텐츠' : '+ 내가 찜한 리스트' }}</span>
            </button>
          </div>
        </div>
      </div>

      <div class="p-6 md:p-10 space-y-6 overflow-y-auto">
        <div class="flex flex-wrap items-center gap-3 text-xs md:text-sm">
          <span class="text-emerald-400 font-bold">{{ title.match_score_pct }}% 일치</span>
          <span class="text-zinc-400">{{ title.release_year }}</span>
          <span class="px-1.5 py-0.5 border border-zinc-600 rounded text-xs text-zinc-300">
            {{ title.maturity_rating }}
          </span>
          <span class="px-1.5 py-0.5 border border-zinc-600 rounded text-xs text-zinc-300">
            {{ title.max_resolution }}
          </span>
          <span v-if="title.is_netflix_original" class="text-red-500 font-semibold text-xs">
            NETFLIX 오리지널
          </span>
        </div>

        <p class="text-sm md:text-base text-zinc-200 leading-relaxed">
          {{ title.localized_synopsis }}
        </p>

        <div class="pt-4 border-t border-zinc-800 grid grid-cols-1 md:grid-cols-2 gap-4 text-xs text-zinc-400">
          <div>
            <span class="text-zinc-500">특징: </span>
            <span class="text-zinc-300">흥미진진, 서스펜스, 현실적</span>
          </div>
          <div>
            <span class="text-zinc-500">오디오 및 자막: </span>
            <span class="text-zinc-300">한국어 [원음], 영어, 한국어 자막</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

