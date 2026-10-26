<!-- clients/web-user/pages/index.vue -->
<script setup lang="ts">
import { ref, onMounted } from 'vue';
import type { AggregatedHomeResponse, LoLoMoItemDto, TitleDto } from '~/types/netflix';
import { useSessionStore } from '~/stores/session';
import { useDefaultCatalog } from '~/composables/useDefaultCatalog';

const sessionStore = useSessionStore();
const config = useRuntimeConfig();
const { getCuratedCatalog, getCuratedTitleDetail, getFallbackPoster } = useDefaultCatalog();

const homeData = ref<AggregatedHomeResponse>(getCuratedCatalog());
const isHeroMuted = ref<boolean>(true);
const activeModalTitle = ref<TitleDto | null>(null);
const activePlaybackTitle = ref<TitleDto | null>(null);
const isAuthModalOpen = ref<boolean>(false);
const authModalMode = ref<'SIGN_IN' | 'SELECT_PROFILE'>('SIGN_IN');

async function loadLoLoMoFeed(): Promise<void> {
  try {
    const res = await $fetch<AggregatedHomeResponse>(
      `${config.public.apiGatewayUrl}/api/v1/browse/lolomo?lang=ko`,
      { headers: sessionStore.authHeaders }
    );
    if (res?.lolomo) homeData.value = res;
  } catch {
    // API Gateway offline; maintain curated catalog view
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

function startPlayback(titleId: string): void {
  activePlaybackTitle.value = getCuratedTitleDetail(titleId);
}

function handleImageError(e: Event): void {
  const el = e.target as HTMLImageElement;
  if (el && el.src !== getFallbackPoster()) {
    el.src = getFallbackPoster();
  }
}

onMounted(() => {
  void loadLoLoMoFeed();
});
</script>

<template>
  <div class="min-h-screen bg-zinc-950 text-white selection:bg-red-600 selection:text-white">
    <NetflixNavbar
      @open-login="authModalMode = 'SIGN_IN'; isAuthModalOpen = true;"
      @open-profiles="authModalMode = sessionStore.isAuthenticated ? 'SELECT_PROFILE' : 'SIGN_IN'; isAuthModalOpen = true;"
      @openLogin="authModalMode = 'SIGN_IN'; isAuthModalOpen = true;"
      @openProfiles="authModalMode = sessionStore.isAuthenticated ? 'SELECT_PROFILE' : 'SIGN_IN'; isAuthModalOpen = true;"
    />

    <section
      v-if="homeData.lolomo.hero_billboard"
      class="relative w-full h-[80vh] min-h-[580px] overflow-hidden bg-black"
    >
      <img
        :src="homeData.lolomo.hero_billboard.personalized_artwork_url"
        alt="Hero Billboard"
        class="w-full h-full object-cover"
        @error="handleImageError"
      />
      <div class="absolute inset-0 bg-gradient-to-r from-black/95 via-black/50 to-transparent" />
      <div class="absolute inset-0 bg-gradient-to-t from-zinc-950 via-zinc-950/60 to-transparent" />

      <div class="absolute inset-0 flex flex-col justify-end pt-28 pb-16 px-6 md:px-16 z-10">
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
              @click="startPlayback(homeData.lolomo.hero_billboard.title_id)"
            >
              ▶ 재생
            </button>
            <button
              class="px-6 py-2.5 bg-zinc-700/80 hover:bg-zinc-600 text-white font-bold text-sm rounded flex items-center gap-2 transition backdrop-blur"
              @click="openTitleDetailModal(homeData.lolomo.hero_billboard.title_id)"
            >
              ⓘ 상세 정보
            </button>
          </div>
        </div>

        <div class="absolute right-6 md:right-16 bottom-16 flex items-center gap-3">
          <button
            class="p-2.5 border border-zinc-400/80 rounded-full hover:border-white transition bg-black/40 text-xs"
            :aria-label="isHeroMuted ? '음소거 해제' : '음소거'"
            @click="isHeroMuted = !isHeroMuted"
          >
            {{ isHeroMuted ? '🔇' : '🔊' }}
          </button>
          <span class="px-2.5 py-1 border-l-2 border-red-600 bg-zinc-900/80 text-xs text-zinc-300 font-bold">
            19+
          </span>
        </div>
      </div>
    </section>

    <main class="-mt-6 md:-mt-10 relative z-20 space-y-10 px-6 md:px-12 pb-20">
      <section v-for="row in homeData.lolomo.rows" :key="row.row_id" class="space-y-3">
        <h2 class="text-base md:text-xl font-bold tracking-tight text-white hover:text-red-500 transition cursor-pointer inline-block">
          {{ row.header_text }} &gt;
        </h2>
        <div class="flex gap-4 overflow-x-auto pb-4 scrollbar-none">
          <article
            v-for="item in row.items"
            :key="item.title_id"
            class="relative flex-none w-44 md:w-60 group cursor-pointer transition-transform duration-300 hover:scale-105"
            @click="openTitleDetailModal(item.title_id)"
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
        <img
          :src="activeModalTitle.hero_artwork_url"
          :alt="activeModalTitle.localized_title"
          class="w-full h-64 md:h-72 object-cover"
          @error="handleImageError"
        />
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
              @click="startPlayback(activeModalTitle.id); activeModalTitle = null;"
            >
              ▶ 지금 재생하기
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

    <NetflixPlayerModal
      :is-open="activePlaybackTitle !== null"
      :title="activePlaybackTitle"
      @close="activePlaybackTitle = null"
    />

    <NetflixAuthModal
      :is-open="isAuthModalOpen"
      :initial-mode="authModalMode"
      @close="isAuthModalOpen = false"
      @authenticated="void loadLoLoMoFeed()"
    />
  </div>
</template>
