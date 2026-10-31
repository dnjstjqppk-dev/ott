<!-- clients/web-user/components/MyListSection.vue -->
<script setup lang="ts">
import type { TitleDto } from '~/types/netflix';

defineProps<{
  items: TitleDto[];
  onOpenDetail: (titleId: string) => void;
  onToggleMyList: (titleId: string) => void;
  myListIds: Set<string>;
  handleImageError: (e: Event) => void;
}>();
</script>

<template>
  <section v-if="items.length" class="px-6 md:px-12 py-6 space-y-3">
    <h3 class="text-lg md:text-xl font-bold text-white">내가 찜한 리스트</h3>
    <div class="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-3">
      <article
        v-for="title in items"
        :key="title.id"
        class="cursor-pointer group"
        @click="onOpenDetail(title.id)"
      >
        <img
          :src="title.hero_artwork_url"
          :alt="title.localized_title"
          class="w-full h-40 object-cover rounded shadow-lg group-hover:scale-105 transition duration-300"
          @error="handleImageError"
        />
        <p class="mt-1 text-xs text-zinc-300 truncate">{{ title.localized_title }}</p>
      </article>
    </div>
  </section>
</template>
