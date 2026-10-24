<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue';
import type { QcReportDetail } from '~/types/studio';

const props = defineProps<{
  videoUrl: string;
  partnerEmail: string;
  partnerIp: string;
  report: QcReportDetail;
  authToken: string;
}>();

const emit = defineEmits<{
  (e: 'decisionSubmitted', updated: QcReportDetail): void;
}>();

const config = useRuntimeConfig();
const videoRef = ref<HTMLVideoElement | null>(null);
const canvasRef = ref<HTMLCanvasElement | null>(null);
const currentMs = ref<number>(0);
const reviewNotes = ref<string>('');
let animFrameId = 0;

const fps = computed<number>(() => {
  const den = Math.max(1, props.report.detected_fps_den);
  return props.report.detected_fps_num / den;
});

const currentFrameNumber = computed<number>(() => {
  return Math.floor((currentMs.value / 1000) * fps.value);
});

function stepSingleFrame(direction: 1 | -1): void {
  if (!videoRef.value) return;
  videoRef.value.pause();
  const frameDurationSec = 1 / fps.value;
  videoRef.value.currentTime = Math.max(0, videoRef.value.currentTime + direction * frameDurationSec);
  currentMs.value = Math.floor(videoRef.value.currentTime * 1000);
}

function seekToIssue(startMs: number): void {
  if (!videoRef.value) return;
  videoRef.value.pause();
  videoRef.value.currentTime = startMs / 1000;
  currentMs.value = startMs;
}

function renderForensicWatermark(): void {
  const canvas = canvasRef.value;
  if (canvas) {
    const ctx = canvas.getContext('2d');
    if (ctx) {
      ctx.clearRect(0, 0, canvas.width, canvas.height);
      ctx.font = 'bold 15px monospace';
      ctx.fillStyle = 'rgba(255, 255, 255, 0.28)';
      const stamp = `${props.partnerEmail} | IP:${props.partnerIp} | F:${currentFrameNumber.value} | ${new Date().toISOString()}`;
      ctx.fillText(stamp, 28, 42);
      ctx.fillText(stamp, 140, canvas.height - 36);
    }
  }
  animFrameId = requestAnimationFrame(renderForensicWatermark);
}

async function submitQcVerdict(verdict: 'PASSED' | 'REJECTED'): Promise<void> {
  const updated = await $fetch<QcReportDetail>(
    `${config.public.apiGatewayUrl}/api/v1/studio/qc/${props.report.id}/decision`,
    {
      method: 'POST',
      headers: { Authorization: `Bearer ${props.authToken}` },
      body: {
        qc_report_id: props.report.id,
        verdict,
        review_notes: reviewNotes.value
      }
    }
  );
  emit('decisionSubmitted', updated);
}

onMounted(() => {
  animFrameId = requestAnimationFrame(renderForensicWatermark);
});

onBeforeUnmount(() => {
  cancelAnimationFrame(animFrameId);
});
</script>

<template>
  <section class="p-6 bg-zinc-900 border border-zinc-800 rounded-xl space-y-5 text-white">
    <div class="flex items-center justify-between">
      <h3 class="text-lg font-bold">QC 프레임 검수 워크벤치 ({{ report.detected_width }}x{{ report.detected_height }} · {{ report.hdr_format }} · {{ report.integrated_lufs }} LUFS)</h3>
      <span class="px-3 py-1 rounded text-xs font-bold bg-amber-500/20 text-amber-300">{{ report.verdict }}</span>
    </div>

    <div class="relative w-full aspect-video bg-black rounded overflow-hidden">
      <video
        ref="videoRef"
        :src="videoUrl"
        class="w-full h-full object-contain"
        controls
        @timeupdate="currentMs = Math.floor(($event.target as HTMLVideoElement).currentTime * 1000)"
      />
      <canvas ref="canvasRef" width="1280" height="720" class="absolute inset-0 w-full h-full pointer-events-none" />
    </div>

    <div class="flex items-center justify-between bg-zinc-950 p-3 rounded border border-zinc-800 text-xs font-mono">
      <div>현재 타임코드: {{ currentMs }} ms | 정확한 프레임: #{{ currentFrameNumber }}</div>
      <div class="flex gap-2">
        <button class="px-3 py-1.5 bg-zinc-800 hover:bg-zinc-700 rounded" @click="stepSingleFrame(-1)">
          -1 프레임 이동
        </button>
        <button class="px-3 py-1.5 bg-zinc-800 hover:bg-zinc-700 rounded" @click="stepSingleFrame(1)">
          +1 프레임 이동
        </button>
      </div>
    </div>

    <div class="space-y-2">
      <h4 class="text-sm font-semibold text-zinc-300">자동 검출 결함 타임라인 ({{ report.issues.length }}건)</h4>
      <div class="max-h-44 overflow-y-auto space-y-1.5">
        <div
          v-for="iss in report.issues"
          :key="iss.id"
          class="flex items-center justify-between p-2.5 bg-zinc-950 border border-zinc-800 rounded text-xs cursor-pointer hover:border-red-500"
          @click="seekToIssue(iss.start_ms)"
        >
          <span class="font-bold text-red-400">[{{ iss.severity }}] {{ iss.issue_type }}</span>
          <span>Frame #{{ iss.start_frame }} ~ #{{ iss.end_frame }} ({{ iss.start_ms }}ms)</span>
          <span class="text-zinc-400">측정치: {{ iss.measured_value }} (기준: {{ iss.threshold_value }})</span>
        </div>
      </div>
    </div>

    <div class="flex gap-3 items-center">
      <input
        v-model="reviewNotes"
        type="text"
        placeholder="검수 의견 및 타임코드 반려 사유 입력..."
        class="flex-1 bg-zinc-950 border border-zinc-700 rounded px-3 py-2 text-sm"
      />
      <button class="px-4 py-2 bg-emerald-600 hover:bg-emerald-500 rounded text-sm font-bold" @click="submitQcVerdict('PASSED')">
        QC 최종 승인
      </button>
      <button class="px-4 py-2 bg-red-600 hover:bg-red-500 rounded text-sm font-bold" @click="submitQcVerdict('REJECTED')">
        반려 (재납품 요청)
      </button>
    </div>
  </section>
</template>

