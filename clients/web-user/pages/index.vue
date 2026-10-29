<!-- clients/web-user/pages/index.vue -->
<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import type { AggregatedHomeResponse, LoLoMoItemDto, TitleDto, PlaybackSessionTicket, EpisodeDto } from '~/types/netflix';
import { useSessionStore } from '~/stores/session';
import { useDefaultCatalog } from '~/composables/useDefaultCatalog';
const sessionStore = useSessionStore();
const config = useRuntimeConfig();
const { getCuratedCatalog, getCuratedTitleDetail, getFallbackPoster } = useDefaultCatalog();
const homeData = ref<AggregatedHomeResponse>(getCuratedCatalog());
const isHeroMuted = ref<boolean>(true);
const activeModalTitle = ref<TitleDto | null>(null);
const activePlaybackTitle = ref<TitleDto | null>(null);
const activePlaybackTicket = ref<PlaybackSessionTicket | null>(null);
const activePlaybackEpisode = ref<EpisodeDto | null>(null);
const isAuthModalOpen = ref<boolean>(false);
const authModalMode = ref<'SIGN_IN' | 'SELECT_PROFILE'>('SIGN_IN');
const isSearchOpen = ref<boolean>(false);
const searchQuery = ref<string>('');

const allCatalogTitles = computed<TitleDto[]>(() => {
  const ids = new Set<string>();
  if (homeData.value.lolomo.hero_billboard) ids.add(homeData.value.lolomo.hero_billboard.title_id);
  for (const row of homeData.value.lolomo.rows) {
    for (const item of row.items) ids.add(item.title_id);
  }
  return Array.from(ids).map((id) => getCuratedTitleDetail(id));
});
const searchResults = computed<TitleDto[]>(() => {
  const q = searchQuery.value.trim().toLowerCase();
  if (!q) return [];
  return allCatalogTitles.value.filter((t) =>
    t.localized_title.toLowerCase().includes(q) ||
    t.localized_synopsis.toLowerCase().includes(q)
  );
});
const continueWatchingItems = computed<LoLoMoItemDto[]>(() => {
  const seen = new Set<string>();
  const items: LoLoMoItemDto[] = [];
  for (const row of homeData.value.lolomo.rows) {
    for (const item of row.items) {
      if (item.bookmark_ms > 0 && !seen.has(item.title_id)) {
        seen.add(item.title_id);
        items.push(item);
      }
    }
  }
  return items;
});
const myListItems = computed<TitleDto[]>(() =>
  Array.from(myListIds.value).map((id) => getCuratedTitleDetail(id))
);
function toggleMyList(titleId: string): void {
  const next = new Set(myListIds.value);
  if (next.has(titleId)) next.delete(titleId);
  else next.add(titleId);
  myListIds.value = next;
  if (import.meta.client) localStorage.setItem('nf-my-list', JSON.stringify(Array.from(next)));
}
function openSearch(): void {
  isSearchOpen.value = true;
}

function closeSearch(): void {
  isSearchOpen.value = false;
  searchQuery.value = '';
}
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

async function startPlayback(titleId: string): Promise<void> {
  const title = getCuratedTitleDetail(titleId);
  activePlaybackTitle.value = title;
  activePlaybackEpisode.value = {
    id: title.id,
    title_id: title.id,
    season_id: '',
    episode_number: 1,
    localized_title: title.localized_title,
    localized_synopsis: title.localized_synopsis,
    duration_ms: 3600000,
    still_image_url: title.hero_artwork_url,
    user_bookmark_ms: 0,
    markers: []
  };
  activePlaybackTicket.value = null;

  if (!sessionStore.isAuthenticated || !sessionStore.deviceId) return;

  try {
    const ticket = await $fetch<PlaybackSessionTicket>(
      `${config.public.apiGatewayUrl}/api/v1/playback/sessions`,
      {
        method: 'POST',
        headers: sessionStore.authHeaders,
        body: {
          episode_id: titleId,
          device_id: sessionStore.deviceId,
          client_ip: '',
          country_code: 'KR'
        }
      }
    );
    activePlaybackTicket.value = ticket;
  } catch {
    activePlaybackTicket.value = null;
  }
}

function handleImageError(e: Event): void {
  const el = e.target as HTMLImageElement;
  if (el && el.src !== getFallbackPoster()) {
    el.src = getFallbackPoster();
  }
}

onMounted(() => {
  if (import.meta.client) {
    try {
      const raw = localStorage.getItem('nf-my-list');
      if (raw) myListIds.value = new Set(JSON.parse(raw));
    } catch {
      myListIds.value = new Set();
    }
  }
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
      @openSearch="openSearch()"
    />

    <HomeHeroSection
      :hero="homeData.lolomo.hero_billboard"
      :is-muted="isHeroMuted"
      :handle-image-error="handleImageError"
      @update:is-muted="isHeroMuted = $event"
      :on-play="() => startPlayback(homeData.lolomo.hero_billboard.title_id)"
      :on-detail="() => openTitleDetailModal(homeData.lolomo.hero_billboard.title_id)"
      :on-toggle-mute="() => isHeroMuted = !isHeroMuted"
    />

    <HomeCatalogRows
      :rows="homeData.lolomo.rows"
      :on-open-detail="openTitleDetailModal"
      :handle-image-error="handleImageError"
    />

    <section v-if="continueWatchingItems.length" class="px-6 md:px-12 py-6 space-y-3">
      <h3 class="text-lg md:text-xl font-bold text-white">?�청 중인 콘텐�?/h3>
      <div class="flex gap-3 overflow-x-auto pb-2">
        <article
          v-for="item in continueWatchingItems"
          :key="item.title_id"
          class="min-w-[180px] md:min-w-[220px] cursor-pointer"
          @click="openTitleDetailModal(item.title_id)"
        >
          <div class="relative rounded overflow-hidden shadow-lg">
            <img
              :src="item.personalized_artwork_url"
              :alt="item.title_id"
              class="w-full h-24 md:h-32 object-cover"
              @error="handleImageError"
            />
            <div class="absolute inset-x-0 bottom-0 h-1 bg-zinc-700">
              <div
                class="h-full bg-red-600"
                :style="{ width: `${Math.min(100, (item.bookmark_ms / item.duration_ms) * 100)}%` }"
              />
            </div>
          </div>
        </article>
      </div>
    </section>

    <MyListSection
      :items="myListItems"
      :on-open-detail="openTitleDetailModal"
      :on-toggle-my-list="toggleMyList"
      :my-list-ids="myListIds"
      :handle-image-error="handleImageError"
    />

    <SearchSection
      :query="searchQuery"
      :results="searchResults"
      :on-open-detail="openTitleDetailModal"
      :on-close="closeSearch"
      :handle-image-error="handleImageError"
    />

    <NetflixPlayerModal
      v-if="!activePlaybackTicket"
      :is-open="activePlaybackTitle !== null"
      :title="activePlaybackTitle"
      :ticket="activePlaybackTicket"
      @close="activePlaybackTitle = null; activePlaybackTicket = null; activePlaybackEpisode = null"
    />

    <div v-else-if="activePlaybackTicket && activePlaybackEpisode" class="fixed inset-0 z-50 bg-black">
      <button
        class="absolute top-4 left-4 z-20 bg-black/70 hover:bg-black text-white rounded-full p-2 text-sm"
        aria-label="?�기"
        @click="activePlaybackTitle = null; activePlaybackTicket = null; activePlaybackEpisode = null"
      >
        ??
      </button>
      <NetflixVideoPlayer
        :episode="activePlaybackEpisode"
        :ticket="activePlaybackTicket"
        @session-revoked="activePlaybackTitle = null; activePlaybackTicket = null; activePlaybackEpisode = null"
      />
    </div>

    <NetflixAuthModal
      :is-open="isAuthModalOpen"
      :initial-mode="authModalMode"
      @close="isAuthModalOpen = false"
      @authenticated="void loadLoLoMoFeed()"
    />
  </div>
</template>
