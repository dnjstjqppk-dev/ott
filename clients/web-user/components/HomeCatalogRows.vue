<!-- clients/web-user/components/HomeCatalogRows.vue -->
<script setup lang="ts">
import type { LoLoMoRowDto } from '~/types/netflix';

defineProps<{
  rows: LoLoMoRowDto[];
  onOpenDetail: (titleId: string) => void;
  handleImageError: (e: Event) => void;
}>();
</script>

<template>
  <main class="-mt-10 md:-mt-16 relative z-20 space-y-10 px-6 md:px-12 pb-20">
    <section v-for="row in rows" :key="row.row_id" class="space-y-3">
      <h2 class="text-base md:text-xl font-bold tracking-tight text-white hover:text-red-500 transition cursor-pointer inline-block">
        {{ row.header_text }} &gt;
      </h2>
      <div class="flex gap-4 overflow-x-auto pb-4 scrollbar-none">
        <article
          v-for="item in row.items"
          :key="item.title_id"
          class="relative flex-none w-44 md:w-60 group cursor-pointer transition-transform duration-300 hover:scale-105"
          @click="onOpenDetail(item.title_id)"
        >
          <div v-if="row.row_type === 'TOP_10_COUNTRY'" class="flex items-center">
            <span
              class="font-black select-none -mr-4 z-0 text-7xl md:text-8xl drop-shadow-[0_2px_4px_rgba(0,0,0,0.9)]"
              style="color: #141414; -webkit-text-stroke: 3px #595959;"
            >
              {{ item.rank_number }}
            </span>
            <img
              :src="item.personalized_artwork_url"
              :alt="item.title_id"
              class="w-32 md:w-44 h-24 md:h-32 object-cover rounded shadow-lg z-10"
              @error="handleImageError"
            />
          </div>
          <div v-else>
            <img
              :src="item.personalized_artwork_url"
              :alt="item.title_id"
              class="w-full h-24 md:h-32 object-cover rounded shadow-lg"
              @error="handleImageError"
            />
            <div v-if="item.duration_ms > 0 && item.bookmark_ms > 0" class="w-full h-1 bg-zinc-700 mt-1 rounded overflow-hidden">
              <div
                class="h-full bg-red-600"
                :style="{ width: `${Math.min(100, (item.bookmark_ms / item.duration_ms) * 100)}%` }"
              />
            </div>
          </div>
          <div class="mt-1 flex items-center justify-between text-xs text-zinc-300">
            <span class="text-emerald-400 font-semibold">{{ item.match_score_pct }}% 일치</span>
            <span class="text-zinc-400">4K HDR</span>
          </div>
        </article>
      </div>
    </section>
  </main>
</template>
