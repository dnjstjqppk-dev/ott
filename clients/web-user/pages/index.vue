<!-- clients/web-user/pages/index.vue -->
<script setup lang="ts">
import { ref, onMounted } from 'vue';
import type { AggregatedHomeResponse, LoLoMoItemDto, TitleDto } from '~/types/netflix';
import { useSessionStore } from '~/stores/session';
import { useDefaultCatalog } from '~/composables/useDefaultCatalog';

const sessionStore = useSessionStore();
const config = useRuntimeConfig();
const { getCuratedCatalog, getCuratedTitleDetail } = useDefaultCatalog();

const homeData = ref<AggregatedHomeResponse>(getCuratedCatalog());
const heroPreviewActive = ref<boolean>(false);
const isHeroMuted = ref<boolean>(true);
const activeModalTitle = ref<TitleDto | null>(null);
const isAuthModalOpen = ref<boolean>(false);
const authModalMode = ref<'SIGN_IN' | 'SELECT_PROFILE'>('SIGN_IN');

async function loadLoLoMoFeed(): Promise<void> {
  try {
    const res = await $fetch<AggregatedHomeResponse>(
      `${config.public.apiGatewayUrl}/api/v1/browse/lolomo?lang=ko`,
      { headers: sessionStore.authHeaders }
    );
    if (res && res.lolomo) {
      homeData.value = res;
    }
  } catch {
    // API Gateway offline; gracefully maintain curated catalog view
  }
}

async function openTitleDetailModal(titleId: string): Promise<void> {
  try {
    const res = await $fetch<{ detail: { title: TitleDto } }>(
      `${config.public.apiGatewayUrl}/api/v1/catalog/titles/${titleId}?lang=ko`,
      { headers: sessionStore.authHeaders }
    );
    activeModalTitle.value = res.detail.title;
  } catch {
    activeModalTitle.value = getCuratedTitleDetail(titleId);
  }
}

async function recordArtworkImpression(item: LoLoMoItemDto): Promise<void> {
  try {
    await $fetch(`${config.public.apiGatewayUrl}/api/v1/reco/bandit-feedback`, {
      method: 'POST',
      headers: sessionStore.authHeaders,
      body: {
        profile_id: sessionStore.activeProfile?.id ?? '',
        artwork_id: item.title_id,
        event_type: 'IMPRESSION'
      }
    });
  } catch {
    // Silent telemetry catch
  }
}

function handleOpenLogin(): void {
  authModalMode.value = 'SIGN_IN';
  isAuthModalOpen.value = true;
}

function handleOpenProfiles(): void {
  authModalMode.value = sessionStore.isAuthenticated ? 'SELECT_PROFILE' : 'SIGN_IN';
  isAuthModalOpen.value = true;
}

onMounted(() => {
  void loadLoLoMoFeed();
  setTimeout(() => {
    heroPreviewActive.value = true;
  }, 1200);
});
</script>

<template>
  <div class="min-h-screen bg-zinc-950 text-white pb-20 selection:bg-red-600 selection:text-white">
    <NetflixNavbar
      @open-login="handleOpenLogin"
      @open-profiles="handleOpenProfiles"
    />

    <section
      v-if="homeData.lolomo.hero_billboard"
      class="relative w-full h-[82vh] overflow-hidden bg-black"
    >
      <img
        :src="homeData.lolomo.hero_billboard.personalized_artwork_url"
        alt="Hero Billboard"
        class="w-full h-full object-cover transition-opacity duration-1000"
      />
      <div class="absolute inset-0 bg-gradient-to-r from-black/90 via-black/40 to-transparent flex flex-col justify-end p-8 md:p-16 pb-28">
        <div class="max-w-2xl space-y-4">
          <div class="flex items-center gap-2">
            <span class="text-red-600 font-black tracking-widest text-base">NETFLIX</span>
            <span class="text-zinc-400 font-semibold text-xs tracking-wider">SERIES</span>
          </div>
          <h1 class="text-4xl md:text-6xl font-black tracking-tight leading-tight">
            오징어 게임: 시즌 2
          </h1>
          <div class="flex items-center gap-3 text-sm">
            <span class="text-emerald-400 font-bold">{{ homeData.lolomo.hero_billboard.match_score_pct }}% 일치</span>
            <span class="px-1.5 py-0.5 border border-zinc-500 text-xs rounded">19+</span>
            <span class="px-1.5 py-0.5 border border-zinc-500 text-xs rounded">{{ homeData.subscription.max_resolution }}</span>
            <span class="text-zinc-300 text-xs">시즌 2개</span>
          </div>
          <p class="text-zinc-200 text-sm md:text-base line-clamp-3 leading-relaxed">
            456억 원의 상금이 걸린 의문의 서바이벌 게임에 다시 뛰어든 기훈. 주최자들의 정체를 파헤치기 위해 목숨을 건 복수가 시작된다.
          </p>
          <div class="flex items-center gap-4 pt-2">
            <button
              class="px-8 py-3 bg-white text-black font-bold text-sm md:text-base rounded flex items-center gap-2 hover:bg-zinc-200 transition shadow"
              @click="openTitleDetailModal(homeData.lolomo.hero_billboard.title_id)"
            >
              ▶ 재생
            </button>
            <button
              class="px-7 py-3 bg-zinc-700/80 hover:bg-zinc-600/90 text-white font-bold text-sm md:text-base rounded flex items-center gap-2 transition backdrop-blur"
              @click="openTitleDetailModal(homeData.lolomo.hero_billboard.title_id)"
            >
              ⓘ 상세 정보
            </button>
            <button
              class="ml-auto p-2 border border-white/60 rounded-full text-xs hover:border-white transition"
              :aria-label="isHeroMuted ? '음소거 해제' : '음소거'"
              @click="isHeroMuted = !isHeroMuted"
            >
              {{ isHeroMuted ? '🔇' : '🔊' }}
            </button>
          </div>
        </div>
      </div>
    </section>

    <main class="-mt-20 relative z-10 space-y-10 px-6 md:px-12">
      <section v-for="row in homeData.lolomo.rows" :key="row.row_id" class="space-y-3">
        <h2 class="text-lg md:text-xl font-bold tracking-tight text-white hover:text-red-500 transition cursor-pointer inline-block">
          {{ row.header_text }} &gt;
        </h2>
        <div class="flex gap-4 overflow-x-auto pb-4 scrollbar-none">
          <article
            v-for="item in row.items"
            :key="item.title_id"
            class="relative flex-none w-48 md:w-64 group cursor-pointer transition-transform duration-300 hover:scale-105"
            @mouseenter="recordArtworkImpression(item)"
            @click="openTitleDetailModal(item.title_id)"
          >
            <div v-if="row.row_type === 'TOP_10_COUNTRY'" class="flex items-center">
              <span class="text-7xl md:text-8xl font-black text-zinc-800 select-none -mr-4 z-0">
                {{ item.rank_number }}
              </span>
              <img :src="item.personalized_artwork_url" :alt="item.title_id" class="w-36 md:w-48 h-28 md:h-36 object-cover rounded shadow z-10" />
            </div>
            <div v-else>
              <img :src="item.personalized_artwork_url" :alt="item.title_id" class="w-full h-28 md:h-36 object-cover rounded shadow" />
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

    <div
      v-if="activeModalTitle"
      class="fixed inset-0 z-50 bg-black/85 backdrop-blur-sm flex items-center justify-center p-4 md:p-6"
      @click.self="activeModalTitle = null"
    >
      <div class="bg-zinc-900 border border-zinc-800 max-w-2xl w-full rounded-xl overflow-hidden shadow-2xl relative">
        <button
          class="absolute top-3 right-3 z-10 bg-black/70 hover:bg-black text-white rounded-full p-2 text-sm"
          aria-label="닫기"
          @click="activeModalTitle = null"
        >
          ✕
        </button>
        <img :src="activeModalTitle.hero_artwork_url" :alt="activeModalTitle.localized_title" class="w-full h-64 md:h-72 object-cover" />
        <div class="p-6 md:p-8 space-y-4">
          <div class="flex items-center justify-between">
            <h3 class="text-2xl font-black">{{ activeModalTitle.localized_title }}</h3>
            <span class="text-emerald-400 font-bold text-sm">{{ activeModalTitle.match_score_pct }}% 일치</span>
          </div>
          <div class="flex items-center gap-3 text-xs text-zinc-300">
            <span class="px-1.5 py-0.5 border border-zinc-600 rounded">{{ activeModalTitle.maturity_rating }}</span>
            <span>출시: {{ activeModalTitle.release_year }}</span>
            <span class="text-zinc-400">{{ activeModalTitle.max_resolution }}</span>
          </div>
          <p class="text-zinc-200 text-sm leading-relaxed">{{ activeModalTitle.localized_synopsis }}</p>
          <div class="pt-4 flex gap-3">
            <button
              class="px-8 py-2.5 bg-red-600 hover:bg-red-700 text-white font-bold rounded text-sm transition"
              @click="activeModalTitle = null"
            >
              지금 재생하기
            </button>
            <button
              class="px-5 py-2.5 bg-zinc-800 hover:bg-zinc-700 text-white rounded text-sm transition"
              @click="activeModalTitle = null"
            >
              + 내가 찜한 리스트
            </button>
          </div>
        </div>
      </div>
    </div>

    <NetflixAuthModal
      :is-open="isAuthModalOpen"
      :initial-mode="authModalMode"
      @close="isAuthModalOpen = false"
      @authenticated="void loadLoLoMoFeed()"
    />
  </div>
</template>
