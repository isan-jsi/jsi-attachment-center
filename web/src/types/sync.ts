export interface SyncStatus {
  is_running: boolean;
  last_sync_at: string | null;
  next_sync_at: string | null;
  documents_synced: number;
  documents_failed: number;
  dlq_count: number;
}

export interface DLQEntry {
  id: number;
  document_id: number;
  error_message: string;
  retry_count: number;
  max_retries: number;
  created_at: string;
  updated_at: string;
  document_title?: string;
}

export interface SyncLogEntry {
  id: number;
  sync_type: string;
  status: string;
  started_at: string;
  completed_at: string | null;
  documents_processed: number;
  documents_failed: number;
  error_message: string | null;
}
