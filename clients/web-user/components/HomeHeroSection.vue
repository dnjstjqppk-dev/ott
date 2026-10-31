<!-- clients/web-user/components/HomeHeroSection.vue -->
<script setup lang="ts">
import type { LoLoMoItemDto } from '~/types/netflix';

defineProps<{
  hero: LoLoMoItemDto;
  isMuted: boolean;
  handleImageError: (e: Event) => void;
  onPlay: () => void;
  onDetail: () => void;
  onToggleMute: () => void;
}>();

const emit = defineEmits<{
  (e: 'update:isMuted', value: boolean): void;
}>();
</script>

<template>
  <section
    v-if="hero"
    class="relative w-full h-[80vh] min-h-[580px] overflow-hidden bg-black"
  >
    <img
      :src="hero.personalized_artwork_url"
      alt="Hero Billboard"
      class="w-full h-full object-cover"
      @error="handleImageError"
    />
    <div class="absolute inset-0 bg-gradient-to-r from-black/95 via-black/40 to-transparent" />
    <div class="absolute inset-0 bg-gradient-to-t from-zinc-950 via-zinc-950/30 to-transparent" />

    <div class="absolute inset-0 flex flex-col justify-end pt-24 pb-28 md:pb-36 px-6 md:px-16 z-10">
      <div class="max-w-xl space-y-4">
        <div class="flex items-center gap-2">
          <span class="text-red-600 font-black tracking-widest text-sm">NETFLIX</span>
          <span class="text-zinc-400 font-semibold text-xs tracking-wider">SERIES</span>
        </div>
        <h1 class="text-3xl md:text-5xl font-black tracking-tight text-white drop-shadow-md">
          오징어 게임: 시즌 2
        </h1>
        <div class="flex items-center gap-3 text-xs md:text-sm">
          <span class="text-emerald-400 font-bold">99% 일치</span>
          <span class="px-1.5 py-0.5 border border-zinc-500 rounded text-xs">19+</span>
          <span class="px-1.5 py-0.5 border border-zinc-500 rounded text-xs">4K Ultra HD</span>
          <span class="text-zinc-300">시즌 2개</span>
        </div>
        <p class="text-zinc-200 text-xs md:text-sm line-clamp-3 leading-relaxed drop-shadow">
          456억 원의 상금이 걸린 의문의 서바이벌 게임에 다시 뛰어든 기훈. 주최자들의 정체를 파헤치기 위해 목숨을 건 복수가 시작된다.
        </p>
        <div class="flex items-center gap-4 pt-2">
          <button
            class="px-7 py-2.5 bg-white text-black font-bold text-sm rounded flex items-center gap-2 hover:bg-zinc-200 transition shadow"
            @click="onPlay"
          >
            ▶ 재생
          </button>
          <button
            class="px-6 py-2.5 bg-zinc-700/80 hover:bg-zinc-600 text-white font-bold text-sm rounded flex items-center gap-2 transition backdrop-blur"
            @click="onDetail"
          >
            ⓘ 상세 정보
          </button>
        </div>
      </div>

      <div class="absolute right-6 md:right-16 bottom-28 md:bottom-36 flex items-center gap-3">
        <button
          class="p-2.5 border border-zinc-400/80 rounded-full hover:border-white transition bg-black/40 text-xs"
          :aria-label="isMuted ? '음소거 해제' : '음소거'"
          @click="emit('update:isMuted', !isMuted)"
        >
          {{ isMuted ? '🔇' : '🔊' }}
        </button>
        <span class="px-2.5 py-1 border-l-2 border-red-600 bg-zinc-900/80 text-xs text-zinc-300 font-bold">
          19+
        </span>
      </div>
    </div>
  </section>
</template>
