<!-- clients/web-user/components/SearchSection.vue -->
<script setup lang="ts">
import type { TitleDto } from '~/types/netflix';

const props = defineProps<{
  query: string;
  results: TitleDto[];
  onOpenDetail: (titleId: string) => void;
  onClose: () => void;
  handleImageError: (e: Event) => void;
}>();
</script>

<template>
  <div
    v-if="query"
    class="fixed inset-0 z-50 bg-black/90 backdrop-blur-sm p-6 md:p-12 overflow-y-auto"
  >
    <div class="max-w-5xl mx-auto space-y-6">
      <div class="flex items-center gap-3">
        <input
          v-model="query"
          type="text"
          placeholder="제목, 사람, 장르 검색"
          class="flex-1 px-4 py-3 bg-zinc-900 border border-zinc-700 rounded text-white focus:outline-none focus:border-red-600"
        />
        <button
          class="px-4 py-3 bg-zinc-800 hover:bg-zinc-700 rounded text-sm"
          @click="onClose"
        >
          닫기
        </button>
      </div>
      <div class="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-3">
        <article
          v-for="title in results"
          :key="title.id"
          class="cursor-pointer"
          @click="onOpenDetail(title.id); onClose();"
        >
          <img
            :src="title.hero_artwork_url"
            :alt="title.localized_title"
            class="w-full h-40 object-cover rounded shadow-lg"
            @error="handleImageError"
          />
          <p class="mt-1 text-xs text-zinc-300 truncate">{{ title.localized_title }}</p>
        </article>
      </div>
      <p v-if="query && !results.length" class="text-zinc-400 text-sm">
        검색 결과가 없습니다.
      </p>
    </div>
  </div>
</template>