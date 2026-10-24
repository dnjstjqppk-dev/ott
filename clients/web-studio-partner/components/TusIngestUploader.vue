<script setup lang="ts">
import { ref, computed } from 'vue';
import type { SourceContainerType } from '~/types/studio';
import { useTusUploader } from '~/composables/useTusUploader';

const props = defineProps<{
  episodeId: string;
  userId: string;
  partnerOrgId: string;
  authToken: string;
}>();

const emit = defineEmits<{
  (e: 'ingestCompleted', qcReportId: string): void;
}>();

const config = useRuntimeConfig();
const selectedFile = ref<File | null>(null);
const containerType = ref<SourceContainerType>('MOV_PRORES');

const {
  isUploading,
  uploadedBytes,
  totalBytes,
  verifiedSha256,
  statusMessage,
  startOrResumeUpload,
  pauseUpload
} = useTusUploader(config.public.apiGatewayUrl, props.authToken);

const progressPct = computed<number>(() => {
  if (totalBytes.value <= 0) return 0;
  return Math.round((uploadedBytes.value / totalBytes.value) * 10000) / 100;
});

function onFileSelected(e: Event): void {
  const input = e.target as HTMLInputElement;
  selectedFile.value = input.files?.[0] ?? null;
}

async function beginUpload(): Promise<void> {
  if (!selectedFile.value) return;
  const result = await startOrResumeUpload({
    file: selectedFile.value,
    episodeId: props.episodeId,
    userId: props.userId,
    partnerOrgId: props.partnerOrgId,
    container: containerType.value
  });
  if (result?.sha256_verified) {
    emit('ingestCompleted', result.triggered_qc_report_id);
  }
}
</script>

<template>
  <section class="p-6 bg-zinc-900 border border-zinc-800 rounded-xl space-y-4 text-white">
    <div class="flex items-center justify-between">
      <h3 class="text-lg font-bold">원본 마스터 인입 (TUS 16MB 청크 이어올리기 & SHA-256 검증)</h3>
      <span class="text-xs px-2.5 py-1 bg-zinc-800 rounded text-zinc-300">{{ statusMessage }}</span>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
      <div>
        <label class="block text-xs text-zinc-400 mb-1">컨테이너 포맷</label>
        <select v-model="containerType" class="w-full bg-zinc-950 border border-zinc-700 rounded px-3 py-2 text-sm">
          <option value="MOV_PRORES">Apple ProRes (.mov)</option>
          <option value="MXF_IMF">SMPTE IMF Package (.mxf)</option>
          <option value="MP4">Master MP4 (.mp4)</option>
          <option value="WAV_ADM">Dolby Atmos ADM BWF (.wav)</option>
        </select>
      </div>
      <div class="md:col-span-2">
        <label class="block text-xs text-zinc-400 mb-1">원본 미디어 파일 선택</label>
        <input
          type="file"
          class="w-full text-sm text-zinc-300 file:mr-4 file:py-2 file:px-4 file:rounded file:border-0 file:bg-red-600 file:text-white"
          @change="onFileSelected"
        />
      </div>
    </div>

    <div class="space-y-1">
      <div class="flex justify-between text-xs text-zinc-400">
        <span>바이트 오프셋: {{ uploadedBytes.toLocaleString() }} / {{ totalBytes.toLocaleString() }} Bytes</span>
        <span>{{ progressPct }}%</span>
      </div>
      <div class="w-full h-2.5 bg-zinc-800 rounded overflow-hidden">
        <div class="h-full bg-emerald-500 transition-all" :style="{ width: `${progressPct}%` }" />
      </div>
    </div>

    <div v-if="verifiedSha256" class="p-3 bg-emerald-950/50 border border-emerald-700 rounded text-xs font-mono">
      SHA-256 Verified: {{ verifiedSha256 }}
    </div>

    <div class="flex gap-3">
      <button
        :disabled="!selectedFile || isUploading"
        class="px-5 py-2 bg-red-600 hover:bg-red-500 disabled:opacity-40 rounded text-sm font-bold"
        @click="beginUpload"
      >
        {{ uploadedBytes > 0 ? '중단 지점부터 이어올리기' : '무결성 인입 시작' }}
      </button>
      <button
        v-if="isUploading"
        class="px-5 py-2 bg-zinc-700 hover:bg-zinc-600 rounded text-sm font-bold"
        @click="pauseUpload"
      >
        일시정지
      </button>
    </div>
  </section>
</template>

