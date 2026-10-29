<!-- clients/web-user/components/NetflixAuthModal.vue -->
<script setup lang="ts">
import { ref, watch } from 'vue';
import type { ProfileDto } from '~/types/netflix';
import { useSessionStore } from '~/stores/session';

const props = defineProps<{ isOpen: boolean; initialMode?: 'SIGN_IN' | 'SELECT_PROFILE' }>();
const emit = defineEmits<{ (e: 'close'): void; (e: 'authenticated'): void }>();

const sessionStore = useSessionStore();
const currentMode = ref<'SIGN_IN' | 'SELECT_PROFILE'>(props.initialMode ?? 'SIGN_IN');
const email = ref<string>('viewer@netflix.com');
const password = ref<string>('netflix2026!strict');
const pinCode = ref<string>('');
const selectedProfileForPin = ref<ProfileDto | null>(null);
const errorMessage = ref<string>('');
const isSubmitting = ref<boolean>(false);

watch(() => props.initialMode, (mode) => {
  if (mode) currentMode.value = mode;
});

watch(() => props.isOpen, (open) => {
  if (open) {
    currentMode.value = props.initialMode ?? (sessionStore.isAuthenticated ? 'SELECT_PROFILE' : 'SIGN_IN');
    selectedProfileForPin.value = null;
    pinCode.value = '';
    errorMessage.value = '';
  }
});



async function handleLogin(): Promise<void> {
  errorMessage.value = '';
  isSubmitting.value = true;
  try {
    await sessionStore.loginWithCredentials(email.value, password.value, 'browser-device-fp-v1');
    currentMode.value = 'SELECT_PROFILE';
  } catch {
    errorMessage.value = '로그인에 실패했습니다. 다시 시도해 주세요.';
  } finally {
    isSubmitting.value = false;
  }
}

async function handleSelectProfile(profile: ProfileDto): Promise<void> {
  if (profile.has_pin) {
    selectedProfileForPin.value = profile;
    return;
  }
  const ok = await sessionStore.selectProfile(profile);
  if (!ok) {
    errorMessage.value = '프로필 선택에 실패했습니다. 다시 시도해 주세요.';
    return;
  }
  emit('authenticated');
  emit('close');
}

async function verifyPinAndComplete(): Promise<void> {
  if (!selectedProfileForPin.value) return;
  if (pinCode.value.length !== 4) {
    errorMessage.value = '잘못된 PIN 번호입니다. 다시 입력해 주세요 (기본: 1234).';
    return;
  }
  isSubmitting.value = true;
  errorMessage.value = '';
  try {
    const ok = await sessionStore.selectProfile(selectedProfileForPin.value, pinCode.value);
    if (!ok) {
      errorMessage.value = '잘못된 PIN 번호입니다. 다시 입력해 주세요 (기본: 1234).';
      return;
    }
    selectedProfileForPin.value = null;
    pinCode.value = '';
    emit('authenticated');
    emit('close');
  } catch {
    errorMessage.value = 'PIN 검증에 실패했습니다. 다시 시도해 주세요.';
  } finally {
    isSubmitting.value = false;
  }
}
</script>

<template>
  <div
    v-if="isOpen"
    class="fixed inset-0 z-50 bg-black/85 backdrop-blur-sm flex items-center justify-center p-4"
    @click.self="$emit('close')"
  >
    <div class="relative bg-zinc-900 border border-zinc-800 rounded-xl p-8 max-w-md w-full shadow-2xl text-white">
      <button
        class="absolute top-4 right-4 text-zinc-400 hover:text-white text-lg p-1"
        aria-label="닫기"
        @click="$emit('close')"
      >
        ✕
      </button>

      <div v-if="selectedProfileForPin" class="space-y-6 text-center">
        <h2 class="text-xl font-bold">프로필 잠금 PIN 입력</h2>
        <p class="text-xs text-zinc-400">
          '{{ selectedProfileForPin.display_name }}' 프로필은 PIN으로 보호되어 있습니다.
        </p>
        <input
          v-model="pinCode"
          type="password"
          maxlength="4"
          placeholder="PIN 4자리"
          class="w-36 mx-auto text-center tracking-widest text-2xl px-4 py-2 bg-black border border-zinc-700 rounded focus:outline-none focus:border-red-600"
          @keyup.enter="verifyPinAndComplete"
        />
        <div v-if="errorMessage" class="text-xs text-red-400">{{ errorMessage }}</div>
        <div class="flex gap-3 justify-center">
          <button
            class="px-5 py-2 bg-red-600 hover:bg-red-700 text-white rounded font-bold text-xs"
            @click="verifyPinAndComplete"
          >
            확인
          </button>
          <button
            class="px-5 py-2 bg-zinc-800 hover:bg-zinc-700 text-white rounded text-xs"
            @click="selectedProfileForPin = null; pinCode = '';"
          >
            취소
          </button>
        </div>
      </div>

      <div v-else-if="currentMode === 'SELECT_PROFILE'" class="space-y-6 text-center">
        <h2 class="text-2xl font-bold">누가 시청 중인가요?</h2>
        <div class="flex flex-wrap justify-center gap-5 pt-2">
          <div
            v-for="profile in sessionStore.profiles"
            :key="profile.id"
            class="flex flex-col items-center gap-2 cursor-pointer group"
            @click="handleSelectProfile(profile)"
          >
            <div
              class="w-20 h-20 rounded-md border-2 border-transparent group-hover:border-white transition flex items-center justify-center font-black text-2xl shadow-lg"
              :class="profile.is_kids ? 'bg-amber-600' : profile.has_pin ? 'bg-emerald-700' : 'bg-red-600'"
            >
              {{ profile.display_name[0] }}
            </div>
            <span class="text-xs text-zinc-300 group-hover:text-white font-medium">
              {{ profile.display_name }}
            </span>
          </div>
        </div>
        <button
          class="text-xs text-zinc-400 hover:text-zinc-200 underline pt-2"
          @click="currentMode = 'SIGN_IN'"
        >
          다른 계정으로 로그인
        </button>
      </div>

      <div v-else class="space-y-5">
        <div>
          <h2 class="text-2xl font-bold">넷플릭스 로그인</h2>
          <p class="text-xs text-zinc-400 mt-1">계정 정보를 입력하거나 원클릭 게스트 로그인을 진행하세요.</p>
        </div>

        <form class="space-y-4" @submit.prevent="handleLogin">
          <div>
            <label class="block text-xs text-zinc-400 mb-1">이메일 주소</label>
            <input
              v-model="email"
              type="email"
              required
              class="w-full px-3 py-2 bg-zinc-950 border border-zinc-700 rounded text-sm text-white focus:outline-none focus:border-red-600"
            />
          </div>
          <div>
            <label class="block text-xs text-zinc-400 mb-1">비밀번호</label>
            <input
              v-model="password"
              type="password"
              required
              class="w-full px-3 py-2 bg-zinc-950 border border-zinc-700 rounded text-sm text-white focus:outline-none focus:border-red-600"
            />
          </div>

          <div v-if="errorMessage" class="text-xs text-red-400">{{ errorMessage }}</div>

          <button
            type="submit"
            :disabled="isSubmitting"
            class="w-full py-2.5 bg-red-600 hover:bg-red-700 text-white font-bold rounded text-sm transition"
          >
            {{ isSubmitting ? '로그인 중...' : '로그인' }}
          </button>
        </form>

        <div class="pt-2 border-t border-zinc-800 text-center">
          <button
            class="text-xs text-zinc-400 hover:text-white transition"
            @click="handleLogin"
          >
            ⚡ 데모 체험 계정으로 바로 시작하기
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

