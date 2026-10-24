<script setup lang="ts">
import { ref, onMounted } from 'vue';
import type {
  BanditArtworkMetric,
  EncodingJobRecord,
  ImmutableAuditLogRecord,
  LicensingWindowRecord,
  OpenConnectEdgeNode
} from '~/types/admin';

const config = useRuntimeConfig();
const adminToken = ref<string>('');
const activeTitleId = ref<string>('00000000-0000-0000-0000-000000000001');
const windows = ref<LicensingWindowRecord[]>([]);
const banditMetrics = ref<BanditArtworkMetric[]>([]);
const encodingJobs = ref<EncodingJobRecord[]>([]);
const edgeNodes = ref<OpenConnectEdgeNode[]>([]);
const auditLogs = ref<ImmutableAuditLogRecord[]>([]);

async function loadControlPlaneSnapshot(): Promise<void> {
  const res = await $fetch<{
    windows: LicensingWindowRecord[];
    bandit_metrics: BanditArtworkMetric[];
    encoding_jobs: EncodingJobRecord[];
    edge_nodes: OpenConnectEdgeNode[];
    audit_logs: ImmutableAuditLogRecord[];
  }>(`${config.public.apiGatewayUrl}/api/v1/admin/control-plane/snapshot`, {
    headers: { Authorization: `Bearer ${adminToken.value}` }
  });
  windows.value = res.windows;
  banditMetrics.value = res.bandit_metrics;
  encodingJobs.value = res.encoding_jobs;
  edgeNodes.value = res.edge_nodes;
  auditLogs.value = res.audit_logs;
}

onMounted(() => {
  void loadControlPlaneSnapshot();
});
</script>

<template>
  <div class="min-h-screen bg-zinc-950 text-white p-8 space-y-8">
    <header class="flex items-center justify-between border-b border-zinc-800 pb-4">
      <div>
        <h1 class="text-2xl font-black text-red-600 tracking-tight">NETFLIX CONTROL PLANE & SRE ADMIN</h1>
        <p class="text-xs text-zinc-400">Catalog Licensing · Media VMAF Pipeline · Open Connect Edge CDN · Immutable Audit Ledger</p>
      </div>
      <button class="px-4 py-2 bg-zinc-800 hover:bg-zinc-700 rounded text-xs font-bold" @click="loadControlPlaneSnapshot">
        실시간 텔레메트리 새로고침
      </button>
    </header>

    <CatalogLicenseCms
      :title-id="activeTitleId"
      :auth-token="adminToken"
      :initial-windows="windows"
      :bandit-metrics="banditMetrics"
    />

    <EncodingAndCdnControl
      :auth-token="adminToken"
      :initial-jobs="encodingJobs"
      :initial-edge-nodes="edgeNodes"
    />

    <section class="p-6 bg-zinc-900 border border-zinc-800 rounded-xl space-y-4">
      <h3 class="text-lg font-bold">불변 보안 감사 로그 (Immutable Audit Ledger · SHA-256 체인 검증)</h3>
      <div class="space-y-2 max-h-80 overflow-y-auto">
        <div
          v-for="log in auditLogs"
          :key="log.id"
          class="p-3 bg-zinc-950 border border-zinc-800 rounded text-xs space-y-1 font-mono"
        >
          <div class="flex justify-between text-zinc-300">
            <span class="font-bold text-amber-400">[{{ log.action_code }}] {{ log.target_entity_type }} ({{ log.target_entity_id }})</span>
            <span>{{ log.created_at }} | Actor: {{ log.actor_role }} ({{ log.actor_ip }})</span>
          </div>
          <div class="grid grid-cols-2 gap-2 text-zinc-400 bg-zinc-900 p-2 rounded">
            <div>Before: {{ JSON.stringify(log.before_state) }}</div>
            <div>After: {{ JSON.stringify(log.after_state) }}</div>
          </div>
          <div class="text-emerald-500 text-[11px]">Record SHA-256: {{ log.record_sha256 }}</div>
        </div>
      </div>
    </section>
  </div>
</template>

