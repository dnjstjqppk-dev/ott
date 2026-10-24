<script setup lang="ts">
import { ref, onMounted } from 'vue';
import type { AggregatedHomeResponse, LoLoMoItemDto, TitleDto } from '~/types/netflix';
import { useSessionStore } from '~/stores/session';

const sessionStore = useSessionStore();
const config = useRuntimeConfig();

const homeData = ref<AggregatedHomeResponse | null>(null);
const heroPreviewActive = ref<boolean>(false);
const isHeroMuted = ref<boolean>(true);
const activeModalTitle = ref<TitleDto | null>(null);

async function loadLoLoMoFeed(): Promise<void> {
  const res = await $fetch<AggregatedHomeResponse>(
    `${config.public.apiGatewayUrl}/api/v1/browse/lolomo?lang=ko`,
    { headers: sessionStore.authHeaders }
  );
  homeData.value = res;
  setTimeout(() => {
    heroPreviewActive.value = true;
  }, 1500);
}

async function openTitleDetailModal(titleId: string): Promise<void> {
  const res = await $fetch<{ detail: { title: TitleDto } }>(
    `${config.public.apiGatewayUrl}/api/v1/catalog/titles/${titleId}?lang=ko`,
    { headers: sessionStore.authHeaders }
  );
  activeModalTitle.value = res.detail.title;
}

async function recordArtworkImpression(item: LoLoMoItemDto): Promise<void> {
  await $fetch(`${config.public.apiGatewayUrl}/api/v1/reco/bandit-feedback`, {
    method: 'POST',
    headers: sessionStore.authHeaders,
    body: {
      profile_id: sessionStore.activeProfile?.id ?? '',
      artwork_id: item.title_id,
      event_type: 'IMPRESSION'
    }
  });
}

onMounted(() => {
  void loadLoLoMoFeed();
});
</script>

<template>
  <div class="min-h-screen bg-zinc-950 text-white pb-20">
    <section
      v-if="homeData?.lolomo.hero_billboard"
      class="relative w-full h-[78vh] overflow-hidden bg-black"
    >
      <img
        v-show="!heroPreviewActive"
        :src="homeData.lolomo.hero_billboard.personalized_artwork_url"
        alt="Hero Billboard"
        class="w-full h-full object-cover"
      />
      <video
        v-if="heroPreviewActive"
        :src="`${config.public.apiGatewayUrl}/preview/${homeData.lolomo.hero_billboard.title_id}.mp4`"
        :muted="isHeroMuted"
        autoplay
        loop
        playsinline
        class="w-full h-full object-cover"
      />
      <div class="absolute inset-0 bg-gradient-to-r from-black/80 via-black/30 to-transparent flex flex-col justify-end p-12 pb-24">
        <span class="text-red-600 font-extrabold tracking-widest text-sm mb-2">NETFLIX ORIGINAL</span>
        <div class="flex items-center gap-3 mb-6">
          <span class="text-emerald-400 font-bold">{{ homeData.lolomo.hero_billboard.match_score_pct }}% 일치</span>
          <span class="px-2 py-0.5 border border-zinc-500 text-xs rounded">{{ homeData.subscription.max_resolution }}</span>
        </div>
        <div class="flex items-center gap-4">
          <NuxtLink
            :to="`/watch/${homeData.lolomo.hero_billboard.title_id}`"
            class="px-8 py-3 bg-white text-black font-bold rounded hover:bg-zinc-200 transition"
          >
            재생
          </NuxtLink>
          <button
            class="px-8 py-3 bg-zinc-700/80 text-white font-bold rounded hover:bg-zinc-600 transition"
            @click="openTitleDetailModal(homeData.lolomo.hero_billboard.title_id)"
          >
            상세 정보
          </button>
          <button
            class="ml-auto px-4 py-2 border border-white/60 rounded-full text-xs"
            @click="isHeroMuted = !isHeroMuted"
          >
            {{ isHeroMuted ? '음소거 해제' : '음소거' }}
          </button>
        </div>
      </div>
    </section>

    <main class="-mt-16 relative z-10 space-y-10 px-12">
      <section v-for="row in homeData?.lolomo.rows" :key="row.row_id" class="space-y-3">
        <h2 class="text-xl font-bold tracking-tight">{{ row.header_text }}</h2>
        <div class="flex gap-4 overflow-x-auto pb-4 scrollbar-none">
          <article
            v-for="item in row.items"
            :key="item.title_id"
            class="relative flex-none w-64 group cursor-pointer transition-transform duration-300 hover:scale-105"
            @mouseenter="recordArtworkImpression(item)"
            @click="openTitleDetailModal(item.title_id)"
          >
            <div v-if="row.row_type === 'TOP_10_COUNTRY'" class="flex items-center">
              <span class="text-8xl font-black text-zinc-800 stroke-white select-none -mr-4 z-0">
                {{ item.rank_number }}
              </span>
              <img :src="item.personalized_artwork_url" :alt="item.title_id" class="w-48 h-36 object-cover rounded z-10" />
            </div>
            <div v-else>
              <img :src="item.personalized_artwork_url" :alt="item.title_id" class="w-full h-36 object-cover rounded" />
              <div v-if="item.duration_ms > 0 && item.bookmark_ms > 0" class="w-full h-1 bg-zinc-700 mt-1 rounded overflow-hidden">
                <div
                  class="h-full bg-red-600"
                  :style="{ width: `${Math.min(100, (item.bookmark_ms / item.duration_ms) * 100)}%` }"
                />
              </div>
            </div>
            <div class="mt-1 flex items-center justify-between text-xs text-zinc-300">
              <span class="text-emerald-400 font-semibold">{{ item.match_score_pct }}% 일치</span>
              <span>4K HDR</span>
            </div>
          </article>
        </div>
      </section>
    </main>

    <div
      v-if="activeModalTitle"
      class="fixed inset-0 z-50 bg-black/80 flex items-center justify-center p-6"
      @click.self="activeModalTitle = null"
    >
      <div class="bg-zinc-900 max-w-3xl w-full rounded-xl overflow-hidden shadow-2xl">
        <img :src="activeModalTitle.hero_artwork_url" :alt="activeModalTitle.localized_title" class="w-full h-72 object-cover" />
        <div class="p-8 space-y-4">
          <div class="flex items-center justify-between">
            <h3 class="text-2xl font-bold">{{ activeModalTitle.localized_title }}</h3>
            <span class="text-emerald-400 font-bold">{{ activeModalTitle.match_score_pct }}% 일치</span>
          </div>
          <p class="text-zinc-300 text-sm leading-relaxed">{{ activeModalTitle.localized_synopsis }}</p>
          <div class="flex items-center gap-3 text-xs text-zinc-400">
            <span>출시: {{ activeModalTitle.release_year }}</span>
            <span>등급: {{ activeModalTitle.maturity_rating }}</span>
            <span>해상도: {{ activeModalTitle.max_resolution }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

