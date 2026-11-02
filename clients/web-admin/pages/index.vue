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
const adminEmail = ref<string>('admin@netflix.com');
const adminPassword = ref<string>('netflix2026!strict');
const loginError = ref<string>('');
const isLoggingIn = ref<boolean>(false);

const activeTitleId = ref<string>('00000000-0000-0000-0000-000000000001');
const windows = ref<LicensingWindowRecord[]>([]);
const banditMetrics = ref<BanditArtworkMetric[]>([]);
const encodingJobs = ref<EncodingJobRecord[]>([]);
const edgeNodes = ref<OpenConnectEdgeNode[]>([]);
const auditLogs = ref<ImmutableAuditLogRecord[]>([]);

async function handleAdminLogin(): Promise<void> {
  loginError.value = '';
  isLoggingIn.value = true;
  try {
    const res = await $fetch<{ token: string }>(
      `${config.public.apiGatewayUrl}/api/v1/admin/auth/login`,
      {
        method: 'POST',
        body: { email: adminEmail.value, password: adminPassword.value }
      }
    );
    adminToken.value = res.token;
    if (import.meta.client) localStorage.setItem('nf_admin_token', res.token);
    await loadControlPlaneSnapshot();
  } catch {
    loginError.value = '인증에 실패했습니다. 관리자 아이디 및 비밀번호를 확인해 주세요.';
  } finally {
    isLoggingIn.value = false;
  }
}

function handleLogout(): void {
  adminToken.value = '';
  if (import.meta.client) localStorage.removeItem('nf_admin_token');
}

async function loadControlPlaneSnapshot(): Promise<void> {
  try {
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
  } catch {
    // offline or session expired
  }
}

onMounted(() => {
  if (import.meta.client) {
    const stored = localStorage.getItem('nf_admin_token');
    if (stored) {
      adminToken.value = stored;
      void loadControlPlaneSnapshot();
    }
  }
});
</script>

<template>
  <div class="min-h-screen bg-zinc-950 text-white p-6 md:p-10 space-y-8 selection:bg-red-600">
    <!-- Admin Login Screen -->
    <div
      v-if="!adminToken"
      class="min-h-[80vh] flex items-center justify-center"
    >
      <div class="w-full max-w-md bg-zinc-900/90 border border-zinc-800 rounded-2xl p-8 shadow-2xl backdrop-blur-md space-y-6">
        <div class="text-center space-y-2">
          <span class="text-red-600 font-black text-3xl tracking-tighter">NETFLIX</span>
          <h2 class="text-xl font-bold tracking-tight text-white">SRE & Control Plane Admin</h2>
          <p class="text-xs text-zinc-400">카탈로그 라이선싱 · 인코딩 큐 · 엣지 CDN 통합 관제탑</p>
        </div>

        <form class="space-y-4 text-xs" @submit.prevent="handleAdminLogin">
          <div>
            <label class="block text-zinc-400 font-semibold mb-1">관리자 이메일 (Admin ID)</label>
            <input
              v-model="adminEmail"
              type="text"
              required
              class="w-full px-4 py-3 bg-zinc-950 border border-zinc-700 rounded-lg text-white focus:outline-none focus:border-red-600 font-mono"
            />
          </div>

          <div>
            <label class="block text-zinc-400 font-semibold mb-1">관리자 비밀번호 (Password)</label>
            <input
              v-model="adminPassword"
              type="password"
              required
              class="w-full px-4 py-3 bg-zinc-950 border border-zinc-700 rounded-lg text-white focus:outline-none focus:border-red-600 font-mono"
            />
          </div>

          <p v-if="loginError" class="text-xs text-red-500 font-medium">
            {{ loginError }}
          </p>

          <button
            type="submit"
            :disabled="isLoggingIn"
            class="w-full py-3.5 bg-red-600 hover:bg-red-700 font-bold rounded-lg transition shadow text-white text-sm cursor-pointer disabled:opacity-50"
          >
            {{ isLoggingIn ? '인증 진행 중...' : '관제탑 보안 로그인 (Unlock Control Plane)' }}
          </button>
        </form>
      </div>
    </div>

    <!-- Admin Dashboard (When Authenticated) -->
    <div v-else class="space-y-8">
      <header class="flex flex-wrap items-center justify-between border-b border-zinc-800 pb-4 gap-4">
        <div>
          <div class="flex items-center gap-3">
            <span class="text-red-600 font-black text-2xl tracking-tighter">NETFLIX</span>
            <span class="px-2 py-0.5 bg-red-600/20 text-red-400 border border-red-500/30 rounded text-xs font-bold">
              SUPER_ADMIN
            </span>
          </div>
          <p class="text-xs text-zinc-400 pt-1">
            Catalog Licensing · Media VMAF Pipeline · Open Connect Edge CDN · Immutable Audit Ledger
          </p>
        </div>
        <div class="flex items-center gap-3">
          <button
            class="px-4 py-2 bg-zinc-800 hover:bg-zinc-700 rounded text-xs font-bold transition"
            @click="loadControlPlaneSnapshot"
          >
            실시간 텔레메트리 새로고침
          </button>
          <button
            class="px-4 py-2 bg-zinc-900 border border-zinc-700 hover:bg-zinc-800 text-zinc-300 rounded text-xs font-bold transition"
            @click="handleLogout"
          >
            로그아웃
          </button>
        </div>
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
  </div>
</template>
