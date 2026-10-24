<script setup lang="ts">
import { ref } from 'vue';
import type { BanditArtworkMetric, LicensingWindowRecord } from '~/types/admin';

const props = defineProps<{
  titleId: string;
  authToken: string;
  initialWindows: LicensingWindowRecord[];
  banditMetrics: BanditArtworkMetric[];
}>();

const config = useRuntimeConfig();
const windows = ref<LicensingWindowRecord[]>([...props.initialWindows]);
const countryCode = ref<string>('KR');
const startsAt = ref<string>('2026-10-01T00:00:00Z');
const endsAt = ref<string>('2027-10-01T00:00:00Z');
const allowDownload = ref<boolean>(true);
const allowAvod = ref<boolean>(true);

async function registerLicenseWindow(): Promise<void> {
  const created = await $fetch<LicensingWindowRecord>(
    `${config.public.apiGatewayUrl}/api/v1/admin/catalog/${props.titleId}/licensing-windows`,
    {
      method: 'POST',
      headers: { Authorization: `Bearer ${props.authToken}` },
      body: {
        title_id: props.titleId,
        country_code: countryCode.value,
        starts_at: startsAt.value,
        ends_at: endsAt.value,
        allow_download: allowDownload.value,
        allow_avod_tier: allowAvod.value
      }
    }
  );
  windows.value.push(created);
}
</script>

<template>
  <section class="p-6 bg-zinc-900 border border-zinc-800 rounded-xl space-y-6 text-white">
    <h3 class="text-lg font-bold">카탈로그 권리 기간 (Licensing Window) & Contextual Bandit 아트워크 관제</h3>

    <div class="grid grid-cols-1 md:grid-cols-5 gap-3 bg-zinc-950 p-4 rounded border border-zinc-800 text-xs">
      <input v-model="countryCode" placeholder="ISO 국가코드 (KR)" class="bg-zinc-900 border border-zinc-700 rounded px-3 py-2" />
      <input v-model="startsAt" placeholder="시작 일시 (RFC3339)" class="bg-zinc-900 border border-zinc-700 rounded px-3 py-2" />
      <input v-model="endsAt" placeholder="종료 일시 (RFC3339)" class="bg-zinc-900 border border-zinc-700 rounded px-3 py-2" />
      <div class="flex items-center gap-3">
        <label class="flex items-center gap-1"><input v-model="allowDownload" type="checkbox" /> 다운로드 허용</label>
        <label class="flex items-center gap-1"><input v-model="allowAvod" type="checkbox" /> 광고형 허용</label>
      </div>
      <button class="bg-red-600 hover:bg-red-500 font-bold rounded px-4 py-2" @click="registerLicenseWindow">
        권리 윈도우 등록
      </button>
    </div>

    <div class="overflow-x-auto">
      <table class="w-full text-left text-xs border-collapse">
        <thead>
          <tr class="border-b border-zinc-800 text-zinc-400">
            <th class="py-2">국가 (ISO-3166)</th>
            <th class="py-2">시작일시</th>
            <th class="py-2">종료일시</th>
            <th class="py-2">오프라인 다운로드</th>
            <th class="py-2">AVOD 노출</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="w in windows" :key="w.id" class="border-b border-zinc-800/60">
            <td class="py-2 font-bold">{{ w.country_code }}</td>
            <td class="py-2 font-mono">{{ w.starts_at }}</td>
            <td class="py-2 font-mono">{{ w.ends_at }}</td>
            <td class="py-2">{{ w.allow_download ? '허용' : '차단' }}</td>
            <td class="py-2">{{ w.allow_avod_tier ? '허용' : '차단' }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="space-y-2">
      <h4 class="text-sm font-semibold text-zinc-300">Contextual Bandit 아트워크 탐색/활용(Thompson Sampling) 실시간 성과</h4>
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div
          v-for="art in banditMetrics"
          :key="art.artwork_id"
          class="p-3 bg-zinc-950 border border-zinc-800 rounded space-y-2 text-xs"
        >
          <img :src="art.storage_url" :alt="art.artwork_id" class="w-full h-28 object-cover rounded" />
          <div class="flex justify-between">
            <span>클러스터 #{{ art.segment_cluster_id }}</span>
            <span class="text-emerald-400 font-bold">CTR: {{ art.ctr_pct }}%</span>
          </div>
          <div class="text-zinc-400 font-mono">
            노출: {{ art.impressions_count.toLocaleString() }} | 재생전환: {{ art.play_starts_count.toLocaleString() }}
          </div>
          <div class="text-zinc-500 font-mono">Beta(α={{ art.alpha_param }}, β={{ art.beta_param }})</div>
        </div>
      </div>
    </div>
  </section>
</template>

