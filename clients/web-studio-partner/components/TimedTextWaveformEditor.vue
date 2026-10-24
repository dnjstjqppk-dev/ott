<script setup lang="ts">
import { ref, onMounted } from 'vue';
import type { EditableTimedTextCue } from '~/types/studio';

const props = defineProps<{
  trackId: string;
  authorUserId: string;
  authToken: string;
  initialCues: EditableTimedTextCue[];
  shotChangeFramesMs: number[];
}>();

const config = useRuntimeConfig();
const waveformCanvas = ref<HTMLCanvasElement | null>(null);
const cues = ref<EditableTimedTextCue[]>([...props.initialCues]);
const saveResultMsg = ref<string>('');

function recalculateCueMetrics(cue: EditableTimedTextCue): void {
  const durationSec = Math.max(0.001, (cue.end_ms - cue.start_ms) / 1000);
  const totalChars = cue.text_lines.reduce((acc, line) => acc + Array.from(line).length, 0);
  const maxLine = cue.text_lines.reduce((max, line) => Math.max(max, Array.from(line).length), 0);
  const cps = Math.round((totalChars / durationSec) * 100) / 100;
  const durationMs = cue.end_ms - cue.start_ms;

  cue.computed_cps = cps;
  cue.max_line_length = maxLine;
  cue.violates_cps = cps > 12.0 || durationMs < 833 || durationMs > 7000;
  cue.violates_shot_change = props.shotChangeFramesMs.some((scMs) => Math.abs(cue.start_ms - scMs) < 83);
}

function addCueRow(): void {
  const lastEnd = cues.value.length > 0 ? (cues.value[cues.value.length - 1]?.end_ms ?? 0) : 0;
  const newCue: EditableTimedTextCue = {
    cue_sequence: cues.value.length + 1,
    start_ms: lastEnd + 200,
    end_ms: lastEnd + 2200,
    text_lines: ['새 자막 대사를 입력하세요'],
    computed_cps: 6.0,
    max_line_length: 12,
    violates_cps: false,
    violates_shot_change: false
  };
  recalculateCueMetrics(newCue);
  cues.value.push(newCue);
}

function drawWaveformAndCues(): void {
  const canvas = waveformCanvas.value;
  if (!canvas) return;
  const ctx = canvas.getContext('2d');
  if (!ctx) return;

  ctx.fillStyle = '#09090b';
  ctx.fillRect(0, 0, canvas.width, canvas.height);
  ctx.strokeStyle = '#10b981';
  ctx.beginPath();
  for (let x = 0; x < canvas.width; x += 3) {
    const amp = Math.sin(x * 0.09) * Math.cos(x * 0.02) * 28;
    ctx.moveTo(x, 45 - amp);
    ctx.lineTo(x, 45 + amp);
  }
  ctx.stroke();
}

async function saveAllCues(): Promise<void> {
  const res = await $fetch<{
    total_cues: number;
    cps_violation_count: number;
    shot_change_violation_count: number;
    generated_webvtt_url: string;
  }>(`${config.public.apiGatewayUrl}/api/v1/studio/timed-text/${props.trackId}/cues`, {
    method: 'PUT',
    headers: { Authorization: `Bearer ${props.authToken}` },
    body: {
      track_id: props.trackId,
      author_user_id: props.authorUserId,
      cues: cues.value.map((c) => ({
        cue_sequence: c.cue_sequence,
        start_ms: c.start_ms,
        end_ms: c.end_ms,
        text_lines: c.text_lines
      }))
    }
  });
  saveResultMsg.value = `WebVTT 생성 완료 (총 ${res.total_cues}개 큐, CPS 경고 ${res.cps_violation_count}건)`;
}

onMounted(() => {
  drawWaveformAndCues();
});
</script>

<template>
  <section class="p-6 bg-zinc-900 border border-zinc-800 rounded-xl space-y-4 text-white">
    <div class="flex items-center justify-between">
      <h3 class="text-lg font-bold">현지화 자막 파형 에디터 (Netflix Timed Text Style Guide 검증기)</h3>
      <div class="flex gap-2">
        <button class="px-3 py-1.5 bg-zinc-800 hover:bg-zinc-700 rounded text-xs font-bold" @click="addCueRow">
          + 자막 큐 추가
        </button>
        <button class="px-4 py-1.5 bg-red-600 hover:bg-red-500 rounded text-xs font-bold" @click="saveAllCues">
          WebVTT 패키징 저장
        </button>
      </div>
    </div>

    <canvas ref="waveformCanvas" width="1100" height="90" class="w-full h-24 rounded border border-zinc-800" />

    <p v-if="saveResultMsg" class="text-xs text-emerald-400 font-semibold">{{ saveResultMsg }}</p>

    <div class="space-y-2 max-h-80 overflow-y-auto">
      <div
        v-for="cue in cues"
        :key="cue.cue_sequence"
        class="grid grid-cols-12 gap-2 items-center p-2.5 bg-zinc-950 border rounded text-xs"
        :class="cue.violates_cps || cue.violates_shot_change ? 'border-red-500' : 'border-zinc-800'"
      >
        <span class="col-span-1 font-mono text-zinc-400">#{{ cue.cue_sequence }}</span>
        <input
          v-model.number="cue.start_ms"
          type="number"
          class="col-span-2 bg-zinc-900 border border-zinc-700 rounded px-2 py-1 font-mono"
          @input="recalculateCueMetrics(cue)"
        />
        <input
          v-model.number="cue.end_ms"
          type="number"
          class="col-span-2 bg-zinc-900 border border-zinc-700 rounded px-2 py-1 font-mono"
          @input="recalculateCueMetrics(cue)"
        />
        <input
          :value="cue.text_lines.join(' / ')"
          type="text"
          class="col-span-4 bg-zinc-900 border border-zinc-700 rounded px-2 py-1"
          @input="
            cue.text_lines = ($event.target as HTMLInputElement).value.split(' / ');
            recalculateCueMetrics(cue);
          "
        />
        <div class="col-span-3 flex items-center justify-end gap-2 font-mono">
          <span :class="cue.violates_cps ? 'text-red-400 font-bold' : 'text-emerald-400'">
            {{ cue.computed_cps }} CPS
          </span>
          <span v-if="cue.violates_shot_change" class="px-1.5 py-0.5 bg-red-600/30 text-red-300 rounded">
            샷체인지 침범
          </span>
        </div>
      </div>
    </div>
  </section>
</template>

