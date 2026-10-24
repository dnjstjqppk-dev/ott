<script setup lang="ts">
import { ref } from 'vue';
import type { EncodingJobRecord, OpenConnectEdgeNode } from '~/types/admin';

const props = defineProps<{
  authToken: string;
  initialJobs: EncodingJobRecord[];
  initialEdgeNodes: OpenConnectEdgeNode[];
}>();

const config = useRuntimeConfig();
const jobs = ref<EncodingJobRecord[]>([...props.initialJobs]);
const edgeNodes = ref<OpenConnectEdgeNode[]>([...props.initialEdgeNodes]);

async function retryDeadLetterJob(jobId: string): Promise<void> {
  const updated = await $fetch<EncodingJobRecord>(
    `${config.public.apiGatewayUrl}/api/v1/admin/media/jobs/${jobId}/retry`,
    {
      method: 'POST',
      headers: { Authorization: `Bearer ${props.authToken}` },
      body: { job_id: jobId, chunk_index: 0 }
    }
  );
  const idx = jobs.value.findIndex((j) => j.job_id === jobId);
  if (idx !== -1) jobs.value[idx] = updated;
}

async function toggleEdgeDrainStatus(node: OpenConnectEdgeNode): Promise<void> {
  const nextStatus = node.status === 'HEALTHY' ? 'DRAINING' : 'HEALTHY';
  const updated = await $fetch<OpenConnectEdgeNode>(
    `${config.public.apiGatewayUrl}/api/v1/admin/cdn/nodes/${node.id}/status`,
    {
      method: 'POST',
      headers: { Authorization: `Bearer ${props.authToken}` },
      body: { edge_node_id: node.id, target_status: nextStatus }
    }
  );
  const idx = edgeNodes.value.findIndex((n) => n.id === node.id);
  if (idx !== -1) edgeNodes.value[idx] = updated;
}
</script>

<template>
  <section class="p-6 bg-zinc-900 border border-zinc-800 rounded-xl space-y-6 text-white">
    <div class="space-y-3">
      <h3 class="text-lg font-bold">Per-Title 미디어 트랜스코딩 큐 & VMAF 품질 관제</h3>
      <div class="space-y-2">
        <div
          v-for="job in jobs"
          :key="job.job_id"
          class="p-4 bg-zinc-950 border border-zinc-800 rounded space-y-3 text-xs"
        >
          <div class="flex items-center justify-between">
            <span class="font-mono font-bold">Job: {{ job.job_id }} (Episode: {{ job.episode_id }})</span>
            <span class="px-2.5 py-0.5 rounded font-bold bg-zinc-800">{{ job.state }} ({{ job.completed_chunks }}/{{ job.total_chunks }} Chunks)</span>
          </div>
          <div class="grid grid-cols-1 md:grid-cols-3 gap-2">
            <div
              v-for="rep in job.representations"
              :key="rep.id"
              class="p-2 bg-zinc-900 rounded border border-zinc-800 font-mono"
            >
              <div>{{ rep.codec }} · {{ rep.width_px }}x{{ rep.height_px }} @ {{ rep.bitrate_kbps }} kbps</div>
              <div class="text-emerald-400">VMAF Mean: {{ rep.harmonic_mean_vmaf }} (Min: {{ rep.min_chunk_vmaf }})</div>
            </div>
          </div>
          <div v-if="job.state === 'FAILED_DLQ'" class="flex items-center justify-between text-red-400">
            <span>DLQ 에러: {{ job.error_message }}</span>
            <button class="px-3 py-1 bg-red-600 text-white rounded font-bold" @click="retryDeadLetterJob(job.job_id)">
              청크 강제 재인코딩
            </button>
          </div>
        </div>
      </div>
    </div>

    <div class="space-y-3">
      <h3 class="text-lg font-bold">Open Connect Edge CDN 실시간 텔레메트리 & 트래픽 우회(Drain) 스위치</h3>
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div
          v-for="node in edgeNodes"
          :key="node.id"
          class="p-4 bg-zinc-950 border rounded space-y-2 text-xs"
          :class="node.status === 'HEALTHY' ? 'border-emerald-700' : 'border-amber-600'"
        >
          <div class="flex justify-between items-center">
            <span class="font-bold font-mono">{{ node.node_code }} ({{ node.region_code }})</span>
            <span class="px-2 py-0.5 rounded font-bold bg-zinc-800">{{ node.status }}</span>
          </div>
          <div class="font-mono text-zinc-300">대역폭: {{ node.current_bandwidth_gbps }} Gbps | 동시 스트림: {{ node.active_streams }}</div>
          <div class="font-mono text-emerald-400">NVMe 캐시 히트율: {{ (node.cache_hit_ratio * 100).toFixed(2) }}%</div>
          <button
            class="w-full py-1.5 rounded font-bold mt-2"
            :class="node.status === 'HEALTHY' ? 'bg-amber-600 hover:bg-amber-500' : 'bg-emerald-600 hover:bg-emerald-500'"
            @click="toggleEdgeDrainStatus(node)"
          >
            {{ node.status === 'HEALTHY' ? '트래픽 Drain 전환 (우회)' : 'HEALTHY 복구 (트래픽 투입)' }}
          </button>
        </div>
      </div>
    </div>
  </section>
</template>

