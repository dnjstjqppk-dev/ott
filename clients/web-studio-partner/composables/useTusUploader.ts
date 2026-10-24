import { ref } from 'vue';
import type {
  AppendChunkResult,
  CompleteTusResult,
  SourceContainerType,
  TusSessionTicket
} from '~/types/studio';

const CHUNK_SIZE_BYTES = 16 * 1024 * 1024;

export function useTusUploader(apiBaseUrl: string, authToken: string) {
  const isUploading = ref<boolean>(false);
  const isPaused = ref<boolean>(false);
  const uploadedBytes = ref<number>(0);
  const totalBytes = ref<number>(0);
  const activeSessionId = ref<string>('');
  const verifiedSha256 = ref<string>('');
  const statusMessage = ref<string>('대기 중');

  async function sha256HexFromBuffer(buffer: ArrayBuffer): Promise<string> {
    const digest = await crypto.subtle.digest('SHA-256', buffer);
    const byteArray = Array.from(new Uint8Array(digest));
    return byteArray.map((b) => b.toString(16).padStart(2, '0')).join('');
  }

  async function computeFullFileSha256(file: File): Promise<string> {
    statusMessage.value = '원본 전체 SHA-256 체크섬 계산 중...';
    const fullBuffer = await file.arrayBuffer();
    return sha256HexFromBuffer(fullBuffer);
  }

  async function startOrResumeUpload(params: {
    file: File;
    episodeId: string;
    userId: string;
    partnerOrgId: string;
    container: SourceContainerType;
  }): Promise<CompleteTusResult | null> {
    isUploading.value = true;
    isPaused.value = false;
    totalBytes.value = params.file.size;

    const expectedSha = await computeFullFileSha256(params.file);
    const ticket = await $fetch<TusSessionTicket>(`${apiBaseUrl}/api/v1/studio/tus/sessions`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${authToken}` },
      body: {
        episode_id: params.episodeId,
        user_id: params.userId,
        partner_org_id: params.partnerOrgId,
        file_name: params.file.name,
        source_container: params.container,
        total_byte_size: params.file.size,
        expected_sha256: expectedSha
      }
    });

    activeSessionId.value = ticket.ingest_session_id;
    uploadedBytes.value = ticket.current_byte_offset;
    return streamRemainingChunks(params.file, ticket.ingest_session_id);
  }

  async function streamRemainingChunks(file: File, sessionId: string): Promise<CompleteTusResult | null> {
    let offset = uploadedBytes.value;
    let chunkIndex = Math.floor(offset / CHUNK_SIZE_BYTES);

    while (offset < file.size) {
      if (isPaused.value) {
        statusMessage.value = `일시정지됨 (${offset.toLocaleString()} bytes 완료)`;
        isUploading.value = false;
        return null;
      }
      const end = Math.min(offset + CHUNK_SIZE_BYTES, file.size);
      const sliceBuffer = await file.slice(offset, end).arrayBuffer();
      const chunkSha = await sha256HexFromBuffer(sliceBuffer);
      const base64Payload = arrayBufferToBase64(sliceBuffer);

      statusMessage.value = `청크 #${chunkIndex} 전송 및 SHA-256 무결성 대조 중...`;
      const res = await $fetch<AppendChunkResult>(
        `${apiBaseUrl}/api/v1/studio/tus/sessions/${sessionId}/chunks`,
        {
          method: 'PATCH',
          headers: { Authorization: `Bearer ${authToken}` },
          body: {
            ingest_session_id: sessionId,
            chunk_index: chunkIndex,
            start_byte: offset,
            payload: base64Payload,
            chunk_sha256: chunkSha
          }
        }
      );
      offset = res.new_byte_offset;
      uploadedBytes.value = offset;
      chunkIndex += 1;
    }

    return finalizeSession(sessionId);
  }

  async function finalizeSession(sessionId: string): Promise<CompleteTusResult> {
    statusMessage.value = 'MinIO 멀티파트 병합 및 서버 전체 SHA-256 검증 중...';
    const done = await $fetch<CompleteTusResult>(
      `${apiBaseUrl}/api/v1/studio/tus/sessions/${sessionId}/complete`,
      {
        method: 'POST',
        headers: { Authorization: `Bearer ${authToken}` }
      }
    );
    verifiedSha256.value = done.computed_sha256;
    isUploading.value = false;
    statusMessage.value = done.sha256_verified ? '인입 완료 (1바이트 무결성 검증 통과)' : '해시 불일치 오류';
    return done;
  }

  function pauseUpload(): void {
    isPaused.value = true;
  }

  return {
    isUploading,
    isPaused,
    uploadedBytes,
    totalBytes,
    activeSessionId,
    verifiedSha256,
    statusMessage,
    startOrResumeUpload,
    pauseUpload
  };
}

function arrayBufferToBase64(buffer: ArrayBuffer): string {
  const bytes = new Uint8Array(buffer);
  let binary = '';
  for (let i = 0; i < bytes.byteLength; i++) {
    binary += String.fromCharCode(bytes[i] ?? 0);
  }
  return btoa(binary);
}

