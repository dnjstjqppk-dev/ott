<!-- clients/web-user/components/NetflixNavbar.vue -->
<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue';
import { useSessionStore } from '~/stores/session';

const emit = defineEmits<{ (e: 'openLogin'): void; (e: 'openProfiles'): void }>();
const sessionStore = useSessionStore();
const isScrolled = ref<boolean>(false);
const showProfileMenu = ref<boolean>(false);
const searchQuery = ref<string>('');
const isSearchOpen = ref<boolean>(false);

function handleScroll(): void {
  isScrolled.value = window.scrollY > 30;
}

function handleLogout(): void {
  sessionStore.accessToken = '';
  sessionStore.activeProfile = null;
  showProfileMenu.value = false;
}

onMounted(() => {
  window.addEventListener('scroll', handleScroll);
});

onBeforeUnmount(() => {
  window.removeEventListener('scroll', handleScroll);
});
</script>

<template>
  <header
    class="fixed top-0 inset-x-0 z-40 transition-colors duration-300 px-6 md:px-12 py-3 flex items-center justify-between"
    :class="isScrolled ? 'bg-zinc-950/95 backdrop-blur-md shadow-lg' : 'bg-gradient-to-b from-black/90 via-black/40 to-transparent'"
  >
    <div class="flex items-center gap-8">
      <NuxtLink to="/" class="flex items-center">
        <span class="text-red-600 font-black text-2xl md:text-3xl tracking-tighter select-none">
          NETFLIX
        </span>
      </NuxtLink>
      <nav class="hidden md:flex items-center gap-5 text-sm font-medium text-zinc-300">
        <NuxtLink to="/" class="text-white hover:text-white transition">홈</NuxtLink>
        <button class="hover:text-zinc-400 transition" @click="$emit('openProfiles')">시리즈</button>
        <button class="hover:text-zinc-400 transition" @click="$emit('openProfiles')">영화</button>
        <button class="hover:text-zinc-400 transition" @click="$emit('openProfiles')">NEW! 요즘 대세 콘텐츠</button>
        <button class="hover:text-zinc-400 transition" @click="$emit('openProfiles')">내가 찜한 리스트</button>
      </nav>
    </div>

    <div class="flex items-center gap-4 text-white text-sm">
      <div class="relative flex items-center">
        <button
          class="p-1 hover:text-zinc-300 transition"
          aria-label="검색"
          @click="isSearchOpen = !isSearchOpen"
        >
          <svg class="w-5 h-5 fill-none stroke-current" stroke-width="2" viewBox="0 0 24 24">
            <circle cx="11" cy="11" r="8" />
            <path d="m21 21-4.3-4.3" />
          </svg>
        </button>
        <input
          v-if="isSearchOpen"
          v-model="searchQuery"
          type="text"
          placeholder="제목, 사람, 장르"
          class="ml-2 px-3 py-1 bg-black/70 border border-zinc-600 rounded text-xs text-white focus:outline-none focus:border-white w-40 md:w-56"
        />
      </div>

      <button
        class="hidden sm:inline text-xs text-zinc-300 hover:text-white transition"
        @click="$emit('openProfiles')"
      >
        키즈
      </button>

      <button class="p-1 hover:text-zinc-300 transition relative" aria-label="알림">
        <svg class="w-5 h-5 fill-none stroke-current" stroke-width="2" viewBox="0 0 24 24">
          <path d="M6 8a6 6 0 0 1 12 0c0 7 3 9 3 9H3s3-2 3-9" />
          <path d="M10.3 21a1.94 1.94 0 0 0 3.4 0" />
        </svg>
      </button>

      <template v-if="sessionStore.isAuthenticated">
        <div class="relative">
          <button
            class="flex items-center gap-2 cursor-pointer focus:outline-none"
            @click="showProfileMenu = !showProfileMenu"
          >
            <div class="w-8 h-8 rounded bg-red-600 font-bold flex items-center justify-center text-xs text-white">
              {{ (sessionStore.activeProfile?.display_name ?? 'N')[0] }}
            </div>
            <svg class="w-3 h-3 text-zinc-400" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M5.23 7.21a.75.75 0 011.06.02L10 11.168l3.71-3.938a.75.75 0 111.08 1.04l-4.25 4.5a.75.75 0 01-1.08 0l-4.25-4.5a.75.75 0 01.02-1.06z" clip-rule="evenodd" />
            </svg>
          </button>

          <div
            v-if="showProfileMenu"
            class="absolute right-0 mt-2 w-48 bg-zinc-900 border border-zinc-800 rounded shadow-2xl py-2 z-50 text-xs text-zinc-200"
          >
            <div class="px-4 py-2 border-b border-zinc-800 font-semibold text-white">
              {{ sessionStore.activeProfile?.display_name ?? '회원님' }}
            </div>
            <button
              class="w-full text-left px-4 py-2 hover:bg-zinc-800 transition"
              @click="$emit('openProfiles'); showProfileMenu = false;"
            >
              프로필 관리
            </button>
            <button
              class="w-full text-left px-4 py-2 hover:bg-zinc-800 transition text-red-400"
              @click="handleLogout"
            >
              넷플릭스에서 로그아웃
            </button>
          </div>
        </div>
      </template>

      <template v-else>
        <button
          class="px-4 py-1.5 bg-red-600 hover:bg-red-700 text-white font-bold text-xs rounded transition shadow"
          @click="$emit('openLogin')"
        >
          로그인
        </button>
      </template>
    </div>
  </header>
</template>

