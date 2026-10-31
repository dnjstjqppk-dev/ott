<!-- clients/web-user/components/ContinueWatchingSection.vue -->
<script setup lang="ts">
import type { LoLoMoItemDto } from '~/types/netflix';

defineProps<{
  items: LoLoMoItemDto[];
  onOpenDetail: (titleId: string) => void;
  handleImageError: (e: Event) => void;
}>();
</script>

<template>
  <section v-if="items.length" class="px-6 md:px-12 py-6 space-y-3">
    <h3 class="text-lg md:text-xl font-bold text-white">시청 중인 콘텐츠</h3>
    <div class="flex gap-3 overflow-x-auto pb-2">
      <article
        v-for="item in items"
        :key="item.title_id"
        class="min-w-[180px] md:min-w-[220px] cursor-pointer"
        @click="onOpenDetail(item.title_id)"
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
</template>
