export type SourceContainerType = 'MOV_PRORES' | 'MXF_IMF' | 'MKV' | 'MP4' | 'WAV_ADM';

export interface CreateTusSessionPayload {
  episode_id: string;
  user_id: string;
  partner_org_id: string;
  file_name: string;
  source_container: SourceContainerType;
  total_byte_size: number;
  expected_sha256: string;
}

export interface TusSessionTicket {
  ingest_session_id: string;
  chunk_size_bytes: number;
  current_byte_offset: number;
  upload_endpoint_url: string;
}

export interface AppendChunkResult {
  new_byte_offset: number;
  etag: string;
}

export interface CompleteTusResult {
  sha256_verified: boolean;
  computed_sha256: string;
  triggered_qc_report_id: string;
}

export interface QcIssueItem {
  id: string;
  issue_type: string;
  severity: 'INFO' | 'WARNING' | 'CRITICAL_BLOCKER';
  start_ms: number;
  end_ms: number;
  start_frame: number;
  end_frame: number;
  measured_value: string;
  threshold_value: string;
  reviewer_comment: string;
  is_resolved: boolean;
}

export interface QcReportDetail {
  id: string;
  ingest_session_id: string;
  episode_id: string;
  verdict: 'PENDING' | 'RUNNING' | 'PASSED' | 'WARNING' | 'REJECTED' | 'OVERRIDDEN_APPROVED';
  detected_width: number;
  detected_height: number;
  detected_fps_num: number;
  detected_fps_den: number;
  color_space: string;
  hdr_format: string;
  audio_channels: string;
  integrated_lufs: number;
  true_peak_dbtp: number;
  issues: QcIssueItem[];
}

export interface EditableTimedTextCue {
  cue_sequence: number;
  start_ms: number;
  end_ms: number;
  text_lines: string[];
  computed_cps: number;
  max_line_length: number;
  violates_cps: boolean;
  violates_shot_change: boolean;
}

export interface PartnerSettlementItem {
  ledger_id: string;
  title_id: string;
  qualified_hours_viewed: number;
  gross_payout_krw: number;
  net_payout_krw: number;
  status: 'CALCULATED' | 'APPROVED' | 'PAID';
}

